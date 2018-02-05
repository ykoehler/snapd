// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package apparmor_test

import (
	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/apparmor"
	"github.com/snapcore/snapd/interfaces/ifacetest"
	"github.com/snapcore/snapd/snap"
)

type specSuite struct {
	iface    *ifacetest.TestInterface
	spec     *apparmor.Specification
	plugInfo *snap.PlugInfo
	plug     *interfaces.ConnectedPlug
	slotInfo *snap.SlotInfo
	slot     *interfaces.ConnectedSlot
}

var _ = Suite(&specSuite{
	iface: &ifacetest.TestInterface{
		InterfaceName: "test",
		AppArmorConnectedPlugCallback: func(spec *apparmor.Specification, plug *interfaces.ConnectedPlug, slot *interfaces.ConnectedSlot) error {
			spec.AddSnippet("connected-plug")
			return nil
		},
		AppArmorConnectedSlotCallback: func(spec *apparmor.Specification, plug *interfaces.ConnectedPlug, slot *interfaces.ConnectedSlot) error {
			spec.AddSnippet("connected-slot")
			return nil
		},
		AppArmorPermanentPlugCallback: func(spec *apparmor.Specification, plug *snap.PlugInfo) error {
			spec.AddSnippet("permanent-plug")
			return nil
		},
		AppArmorPermanentSlotCallback: func(spec *apparmor.Specification, slot *snap.SlotInfo) error {
			spec.AddSnippet("permanent-slot")
			return nil
		},
	},
	plugInfo: &snap.PlugInfo{
		Snap:      &snap.Info{SuggestedName: "snap1"},
		Name:      "name",
		Interface: "test",
		Apps: map[string]*snap.AppInfo{
			"app1": {
				Snap: &snap.Info{
					SuggestedName: "snap1",
				},
				Name: "app1"}},
	},
	slotInfo: &snap.SlotInfo{
		Snap:      &snap.Info{SuggestedName: "snap2"},
		Name:      "name",
		Interface: "test",
		Apps: map[string]*snap.AppInfo{
			"app2": {
				Snap: &snap.Info{
					SuggestedName: "snap2",
				},
				Name: "app2"}},
	},
})

func (s *specSuite) SetUpTest(c *C) {
	s.spec = &apparmor.Specification{}
	s.plug = interfaces.NewConnectedPlug(s.plugInfo, nil)
	s.slot = interfaces.NewConnectedSlot(s.slotInfo, nil)
}

// The spec.Specification can be used through the interfaces.Specification interface
func (s *specSuite) TestSpecificationIface(c *C) {
	var r interfaces.Specification = s.spec
	c.Assert(r.AddConnectedPlug(s.iface, s.plug, s.slot), IsNil)
	c.Assert(r.AddConnectedSlot(s.iface, s.plug, s.slot), IsNil)
	c.Assert(r.AddPermanentPlug(s.iface, s.plugInfo), IsNil)
	c.Assert(r.AddPermanentSlot(s.iface, s.slotInfo), IsNil)
	c.Assert(s.spec.Snippets(), DeepEquals, map[string][]string{
		"snap.snap1.app1": {"connected-plug", "permanent-plug"},
		"snap.snap2.app2": {"connected-slot", "permanent-slot"},
	})
}

// AddSnippet adds a snippet for the given security tag.
func (s *specSuite) TestAddSnippet(c *C) {
	restore := apparmor.SetSpecScope(s.spec, []string{"snap.demo.demo", "snap.demo.service"}, "demo")
	defer restore()

	// Add two snippets in the context we are in.
	s.spec.AddSnippet("snippet 1")
	s.spec.AddSnippet("snippet 2")

	// The snippets were recorded correctly.
	c.Assert(s.spec.Snippets(), DeepEquals, map[string][]string{
		"snap.demo.demo":    {"snippet 1", "snippet 2"},
		"snap.demo.service": {"snippet 1", "snippet 2"},
	})
	c.Assert(s.spec.SnippetForTag("snap.demo.demo"), Equals, "snippet 1\nsnippet 2")
	c.Assert(s.spec.SecurityTags(), DeepEquals, []string{"snap.demo.demo", "snap.demo.service"})
}

// AddSunSnippet adds a snippet for the snap-update-ns profile for a given snap.
func (s *specSuite) TestAddSunSnippet(c *C) {
	restore := apparmor.SetSpecScope(s.spec, []string{"snap.demo.demo", "snap.demo.service"}, "demo")
	defer restore()

	// Add a two snap-update-ns snippets in the context we are in.
	s.spec.AddSunSnippet("snippet 1")
	s.spec.AddSunSnippet("snippet 2")

	// The snippets were recorded correctly.
	c.Assert(s.spec.SunSnippets(), DeepEquals, map[string][]string{
		"demo": {"snippet 1", "snippet 2"},
	})
}
