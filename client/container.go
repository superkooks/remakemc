package client

import (
	"remakemc/client/renderers"
	"remakemc/core"
	"remakemc/core/proto"

	"github.com/go-gl/glfw/v3.2/glfw"
)

func ProcessContainerInteraction(inventoryHover, hotbarHover int, inventory [27]core.ItemStack, hotbar [9]core.ItemStack, floating core.ItemStack) {
	if inventoryHover < 0 && hotbarHover < 0 {
		return
	}

	// Left click
	if renderers.Win.GetMouseButton(glfw.MouseButton1) == glfw.Press && mouseOne.Invoke() {
		if hotbarHover >= 0 {
			// Hotbar
			if floating.IsEmpty() && !hotbar[hotbarHover].IsEmpty() {
				// Take the stack from the slot
				floating = hotbar[hotbarHover]
				hotbar[hotbarHover] = core.ItemStack{}

			} else if !floating.IsEmpty() && hotbar[hotbarHover].IsEmpty() {
				// Place the stack in the slot
				hotbar[hotbarHover] = floating
				floating = core.ItemStack{}

			} else if !floating.IsEmpty() && !hotbar[hotbarHover].IsEmpty() {
				if floating.Item == hotbar[hotbarHover].Item {
					// Add to stack
					h := hotbar[hotbarHover]
					mss := core.ItemRegistry[h.Item].MaxStackSize
					if h.Count+floating.Count > mss {
						diff := h.Count + floating.Count - mss
						hotbar[hotbarHover].Count = mss
						floating.Count -= diff
					} else {
						hotbar[hotbarHover].Count += floating.Count
						floating = core.ItemStack{}
					}

				} else {
					// Switch items
					hotbar[hotbarHover], floating = floating, hotbar[hotbarHover]
				}
			}

		} else {
			// Inventory
			if floating.IsEmpty() && !inventory[inventoryHover].IsEmpty() {
				// Take the stack from the slot
				floating = inventory[inventoryHover]
				inventory[inventoryHover] = core.ItemStack{}

			} else if !floating.IsEmpty() && inventory[inventoryHover].IsEmpty() {
				// Place the stack in the slot
				inventory[inventoryHover] = floating
				floating = core.ItemStack{}

			} else if !floating.IsEmpty() && !inventory[inventoryHover].IsEmpty() {
				if floating.Item == inventory[inventoryHover].Item {
					// Add to stack
					h := inventory[inventoryHover]
					mss := core.ItemRegistry[h.Item].MaxStackSize
					if h.Count+floating.Count > mss {
						diff := h.Count + floating.Count - mss
						inventory[inventoryHover].Count = mss
						floating.Count -= diff
					} else {
						inventory[inventoryHover].Count += floating.Count
						floating = core.ItemStack{}
					}

				} else {
					// Switch items
					inventory[inventoryHover], floating = floating, inventory[inventoryHover]
				}
			}
		}

		serverWrite <- proto.CONTAINER_CONTENTS
		serverWrite <- proto.ContainerContents{
			Slots:         append(hotbar[:], inventory[:]...),
			FloatingStack: floating,
		}
	}

	// Right click
	if renderers.Win.GetMouseButton(glfw.MouseButton2) == glfw.Press && mouseTwo.Invoke() {
		if hotbarHover >= 0 {
			// Hotbar
			if floating.IsEmpty() && !hotbar[hotbarHover].IsEmpty() {
				// Take the greater half of the stack from the slot
				diff := hotbar[hotbarHover].Count / 2
				floating = hotbar[hotbarHover]
				floating.Count -= diff
				hotbar[hotbarHover].Count = diff

			} else if !floating.IsEmpty() && hotbar[hotbarHover].IsEmpty() {
				// Place items one by one into slot
				hotbar[hotbarHover] = floating
				hotbar[hotbarHover].Count = 1
				floating.Count -= 1

			} else if !floating.IsEmpty() && !hotbar[hotbarHover].IsEmpty() {
				if floating.Item == hotbar[hotbarHover].Item {
					// Place items onee by one in to slot
					hotbar[hotbarHover].Count++
					floating.Count--

				} else {
					// Switch items
					hotbar[hotbarHover], floating = floating, hotbar[hotbarHover]
				}
			}

		} else {
			// Inventory
			if floating.IsEmpty() && !inventory[inventoryHover].IsEmpty() {
				// Take the greater half of the stack from the slot
				diff := inventory[inventoryHover].Count / 2
				floating = inventory[inventoryHover]
				floating.Count -= diff
				inventory[inventoryHover].Count = diff

			} else if !floating.IsEmpty() && inventory[inventoryHover].IsEmpty() {
				// Place items one by one into slot
				inventory[inventoryHover] = floating
				inventory[inventoryHover].Count = 1
				floating.Count -= 1

			} else if !floating.IsEmpty() && !inventory[inventoryHover].IsEmpty() {
				if floating.Item == inventory[inventoryHover].Item {
					// Place items onee by one in to slot
					inventory[inventoryHover].Count++
					floating.Count--

				} else {
					// Switch items
					inventory[inventoryHover], floating = floating, inventory[inventoryHover]
				}
			}
		}

		serverWrite <- proto.CONTAINER_CONTENTS
		serverWrite <- proto.ContainerContents{
			Slots:         append(hotbar[:], inventory[:]...),
			FloatingStack: floating,
		}
	}

	// TODO painting mode
}
