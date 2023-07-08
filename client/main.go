package client

import (
	"fmt"
	"math/rand"
	"net"
	"remakemc/client/gui"
	"remakemc/client/renderers"
	"remakemc/config"
	"remakemc/core"
	"remakemc/core/blocks"
	_ "remakemc/core/entities"
	_ "remakemc/core/items"
	"remakemc/core/proto"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var serverRead chan interface{}
var serverWrite chan interface{}
var conn *net.TCPConn
var player *Player

type meshDone struct {
	position    core.Vec3
	mesh        []float32
	normals     []float32
	uvs         []float32
	lightLevels []float32
}

func Start() {
	runtime.LockOSThread()

	// Initialize texture atlas
	blocks.Grass.RenderType.Init()
	blocks.Dirt.RenderType.Init()
	blocks.Stone.RenderType.Init()
	blocks.Cobblestone.RenderType.Init()

	// Initialize OpenGL and GLFW
	renderers.InitAll(config.App.Client.DefaultWidth, config.App.Client.DefaultHeight)
	defer glfw.Terminate()

	// Initialize gui elements
	gui.Init()
	gui.ReadOTF()

	// Join server
	var err error
	conn, err = net.DialTCP("tcp4", nil, &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 53785,
	})
	if err != nil {
		panic(err)
	}
	serverRead = make(chan interface{}, 32)
	serverWrite = make(chan interface{}, 32)

	// Read and write from network in separate thread
	go readFromNet(serverRead)
	go writeFromQueue(serverWrite)

	// Send join event
	serverWrite <- proto.JOIN
	serverWrite <- proto.Join{
		Username: fmt.Sprintf("test-%v", rand.Intn(100)),
	}

	// Read play event
	msg := (<-serverRead).(proto.Play)

	// Initialize terrain
	chunks := msg.InitialChunks.GetChunks()
	dim := &core.Dimension{
		Lock:   new(sync.RWMutex),
		Chunks: make(map[core.Vec3]*core.Chunk),
	}
	for _, v := range chunks {
		dim.Chunks[v.Position] = v
	}

	t := time.Now()
	for _, v := range dim.Chunks {
		renderers.MakeChunkMeshAndVAO(dim, v)
	}
	fmt.Println("Generated initial terrain VAOs in", time.Since(t))

	// Initialize player
	player = NewPlayer(msg.Player.Position, msg.Player.EntityID)
	player.LookAzimuth = msg.Player.LookAzimuth
	player.LookElevation = msg.Player.LookElevation
	player.Yaw = msg.Player.Yaw
	player.Inventory = msg.Inventory
	player.Hotbar = msg.Hotbar
	renderers.Win.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	renderers.Win.SetScrollCallback(player.ScrollCallback)

	// Initialize inputs
	mouseOne := new(core.Debounced)
	mouseTwo := new(core.Debounced)

	// Get all tickers
	var allTickers []core.Tickable
	allTickers = append(allTickers, player)

	gl.DebugMessageCallback(func(
		source uint32,
		gltype uint32,
		id uint32,
		severity uint32,
		length int32,
		message string,
		userParam unsafe.Pointer) {

		if severity == gl.DEBUG_SEVERITY_NOTIFICATION {
			return
		}

		fmt.Println(message)
	}, nil)

	// Game loop
	previousTime := glfw.GetTime()
	var frames int
	var cumulativeTime float64
	var collectedDelta float64
	for !renderers.Win.ShouldClose() {
		// Get delta time
		windowTime := glfw.GetTime()
		deltaTime := windowTime - previousTime
		previousTime = windowTime

		// Calculate frame rate
		frames++
		cumulativeTime += deltaTime
		if frames == 60 {
			fmt.Println(1/(cumulativeTime/60), "fps")
			frames = 0
			cumulativeTime = 0
		}

		// Clear window
		gl.ClearColor(79.0/255, 167.0/255, 234.0/255, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.Enable(gl.DEBUG_OUTPUT)

		// Process user input and recalculate view matrix
		if renderers.IsWindowFocused() && !inventoryOpen {
			player.ProcessMousePosition(deltaTime)
		}
		player.DoUpdate(deltaTime, dim)
		view := mgl32.LookAtV(
			player.CameraPos(),                       // Camera is at ... in World Space
			player.CameraPos().Add(player.LookDir()), // and looks at
			mgl32.Vec3{0, 1, 0},                      // Head is up
		)

		// Update entities
		for _, v := range dim.Entities {
			v.DoUpdate(deltaTime, dim)
		}

		// See if we need to do a game tick
		collectedDelta += deltaTime
		for ; collectedDelta >= 1.0/20; collectedDelta -= 1.0 / 20 {
			for _, v := range allTickers {
				v.DoTick()
			}
		}

		// Mining
		if !inventoryOpen && renderers.Win.GetMouseButton(glfw.MouseButton1) == glfw.Press && mouseOne.Invoke() {
			core.TraceRay(player.LookDir(), player.CameraPos(), 16, func(v, h mgl32.Vec3) (stop bool) {
				block := dim.GetBlockAt(core.NewVec3FromFloat(v))
				if block.Type != nil {
					serverWrite <- proto.BLOCK_DIG
					serverWrite <- proto.BlockDig{
						Position:      block.Position,
						SubvoxelHit:   h,
						FinishDigging: true, // TODO implement survival digging
					}

					return true
				} else {
					return false
				}
			})
		} else if renderers.Win.GetMouseButton(glfw.MouseButton1) == glfw.Release {
			mouseOne.Reset()
		}

		// Placing
		if !inventoryOpen && renderers.Win.GetMouseButton(glfw.MouseButton2) == glfw.Press && mouseTwo.Invoke() {
			core.TraceRay(player.LookDir(), player.CameraPos(), 16, func(v, h mgl32.Vec3) (stop bool) {
				block := dim.GetBlockAt(core.NewVec3FromFloat(v))
				if block.Type != nil {
					serverWrite <- proto.BLOCK_INTERACTION
					serverWrite <- proto.BlockInteraction{
						Position:    block.Position,
						SubvoxelHit: h,
					}

					return true
				} else {
					return false
				}
			})
		} else if renderers.Win.GetMouseButton(glfw.MouseButton2) == glfw.Release {
			mouseTwo.Reset()
		}

		// Render terrain
		renderers.RenderChunks(dim, view, player.Position)

		// Find selector position and render
		core.TraceRay(player.LookDir(), player.CameraPos(), 16, func(v, _ mgl32.Vec3) (stop bool) {
			block := dim.GetBlockAt(core.NewVec3FromFloat(v))
			if block.Type != nil {
				renderers.RenderSelector(core.NewVec3FromFloat(v).ToFloat(), view)
				return true
			} else {
				return false
			}
		})

		// Render all entities
		for _, v := range dim.Entities {
			core.EntityRegistry[v.EntityType].RenderType.RenderEntity(v, view)
		}

		// Render gui
		gui.RenderGame(player.SelectedHotbarSlot, player.Hotbar)

		gui.RenderText(
			mgl32.Vec2{1, 1},
			fmt.Sprintf("%0.1f fps", 1/(cumulativeTime/float64(frames))),
			gui.Anchor{Vertical: 1, Horizontal: 1},
		)

		if inventoryOpen {
			gui.RenderInventory(player.Inventory, player.Hotbar)
		}

		// Update window
		glfw.PollEvents()
		renderers.Win.SwapBuffers()

		// Read from network
	outer:
		for {
			select {
			case m := <-serverRead:
				switch msg := m.(type) {
				case proto.UnloadChunks:
					dim.Lock.Lock()
					for _, v := range msg {
						fmt.Println("unloading chunk", v)
						renderers.FreeChunk(dim.Chunks[v])
						delete(dim.Chunks, v)
					}
					dim.Lock.Unlock()

				case proto.LoadChunks:
					dim.Lock.Lock()
					chunks := msg.GetChunks()
					for _, v := range chunks {
						dim.Chunks[v.Position] = v
					}
					dim.Lock.Unlock()
					for _, v := range chunks {
						go func(c *core.Chunk) {
							dim.Lock.RLock()
							mesh, normals, uvs, lightLevels := renderers.MakeChunkMesh(dim, c.Position)
							dim.Lock.RUnlock()
							serverRead <- meshDone{position: c.Position, mesh: mesh, normals: normals, uvs: uvs, lightLevels: lightLevels}
						}(v)
					}

				// Special internal event used to transfer mesh data generated in thread
				case meshDone:
					c := dim.Chunks[msg.position]
					renderers.MakeChunkVAO(c, msg.mesh, msg.normals, msg.uvs, msg.lightLevels)

				case proto.EntityCreate:
					e := &core.Entity{
						ID:            msg.EntityID,
						Position:      msg.Position,
						AABB:          msg.AABB,
						EntityType:    msg.EntityType,
						Lerp:          true,
						Yaw:           msg.Yaw,
						Pitch:         msg.Pitch,
						LookAzimuth:   msg.LookAzimuth,
						LookElevation: msg.LookElevation,
					}
					e.NewLerp(msg.Position)
					dim.Entities = append(dim.Entities, e)

				case proto.EntityPosition:
					if msg.EntityID == player.ID {
						// Update the player's position absolutely.
						// Only happens when server thinks divergence is too high.
						// Will cause a rubberband
						player.Position = msg.Position
					} else {
						// Find the entity by ID
						for _, v := range dim.Entities {
							if v.ID == msg.EntityID {
								// Update
								v.LookAzimuth = msg.LookAzimuth
								v.LookElevation = msg.LookElevation
								v.Position = msg.Position
								v.Yaw = msg.Yaw

								v.NewLerp(msg.Position)
							}
						}
					}

				case proto.BlockUpdate:
					dim.Lock.Lock()

					dim.SetBlockAt(core.Block{
						Position: msg.Position,
						Type:     core.BlockRegistry[msg.BlockType],
					})
					renderers.UpdateRequiredMeshes(dim, msg.Position)

					dim.Lock.Unlock()

				case proto.EntityEquipment:
					fmt.Println(msg.HeldItemType)

				case proto.PlayerInventory:
					player.Hotbar = msg.Hotbar
					player.Inventory = msg.Inventory

				}
			default:
				break outer
			}
		}
	}
}
