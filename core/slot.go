package core

import (
	"github.com/go-gl/mathgl/mgl32"
)

// InventorySlot (CraftIn) (CraftOut)
// TempSlot (CraftIn)
// ResultSlot (CraftOut)
// FilteredSlot
// CreativeSlot

type Slot interface {
	// Get the box of the slot in OpenGL coordinates
	GetBox() (start, end mgl32.Vec2)

	// Get the stack currently held by the slot
	GetStack() ItemStack

	// Set the stack
	SetStack(ItemStack)

	// Attempt to take the currently held stack.
	TakeStack(half bool) (stack ItemStack, allowed bool)

	// Attempt to put down a stack, returns unused items
	PutStack(stack ItemStack) (returned ItemStack)

	// Does this slot store items?
	// Or is it like a crafting table, which drops items.
	Temp() bool
}

type InventorySlot struct {
	Stack ItemStack
	Start mgl32.Vec2
	End   mgl32.Vec2
}

func (s *InventorySlot) GetBox() (start, end mgl32.Vec2) {
	return s.Start, s.End
}

func (s *InventorySlot) GetStack() ItemStack {
	return s.Stack
}

func (s *InventorySlot) SetStack(i ItemStack) {
	s.Stack = i
}

func (s *InventorySlot) TakeStack(half bool) (ItemStack, bool) {
	if half {
		// Take half the stack
		m := s.Stack
		s.Stack.Count /= 2
		m.Count -= s.Stack.Count
		return m, true
	} else {
		// Take the entire stack
		m := s.Stack
		s.Stack = ItemStack{}
		return m, true
	}
}

func (s *InventorySlot) PutStack(i ItemStack) ItemStack {
	if s.Stack.IsEmpty() {
		// Put the stack in the empty slot
		s.Stack = i
		return ItemStack{}
	} else if i.Item == s.Stack.Item {
		// Merge the stacks, up to max stack size
		mss := ItemRegistry[i.Item].MaxStackSize
		if i.Count+s.Stack.Count > mss {
			diff := s.Stack.Count + i.Count - mss
			s.Stack.Count = mss
			i.Count = diff
			return i
		} else {
			s.Stack.Count += i.Count
			return ItemStack{}
		}
	} else {
		// Exchange the slots
		m := s.Stack
		s.Stack = i
		return m
	}
}

func (s *InventorySlot) Temp() bool {
	return false
}
