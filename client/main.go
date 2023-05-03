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
	"remakemc/core/proto"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/vmihailenco/msgpack/v5"
)

var serverRead chan interface{}
var serverWrite *msgpack.Encoder
var conn *net.TCPConn

type meshDone struct {
	position    core.Vec3
	mesh        []float32
	normals     []float32
	uvs         []float32
	lightLevels []float32
}

func Start() {
	runtime.LockOSThread()

	// Don't let allocated memory exceed 125% of in-use memory
	debug.SetGCPercent(25)

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
	serverWrite = msgpack.NewEncoder(conn)

	// Send join event
	err = serverWrite.Encode(proto.JOIN)
	if err != nil {
		panic(err)
	}
	err = serverWrite.Encode(proto.Join{
		Username: fmt.Sprintf("test-%v", rand.Intn(100)),
	})
	if err != nil {
		panic(err)
	}

	// Read from network in separate thread
	go func() {
		d := msgpack.NewDecoder(conn)
		for {
			var msgType int
			err := d.Decode(&msgType)
			if err != nil {
				panic(err)
			}

			switch msgType {
			case proto.PLAY:
				var data proto.Play
				err = d.Decode(&data)
				if err != nil {
					panic(err)
				}
				serverRead <- data

			case proto.LOAD_CHUNKS:
				var data proto.LoadChunks
				err = d.Decode(&data)
				if err != nil {
					panic(err)
				}
				serverRead <- data

			case proto.UNLOAD_CHUNKS:
				var data proto.UnloadChunks
				err = d.Decode(&data)
				if err != nil {
					panic(err)
				}
				serverRead <- data

			case proto.ENTITY_CREATE:
				fmt.Println("received entity create")

				var data proto.EntityCreate
				err = d.Decode(&data)
				if err != nil {
					panic(err)
				}
				serverRead <- data

			case proto.ENTITY_POSITION:
				var data proto.EntityPosition
				err = d.Decode(&data)
				if err != nil {
					panic(err)
				}
				serverRead <- data

			default:
				panic("unknown packet type")
			}
		}
	}()

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
	player := NewPlayer(msg.Player.Position, msg.Player.EntityID)
	player.LookAzimuth = msg.Player.LookAzimuth
	player.LookElevation = msg.Player.LookElevation
	player.Yaw = msg.Player.Yaw
	renderers.Win.SetInputMode(glfw.CursorMode, glfw.CursorHidden)

	// Initialize inputs
	mouseOne := new(core.Debounced)
	mouseTwo := new(core.Debounced)

	// Get all tickers
	var allTickers []core.Tickable
	allTickers = append(allTickers, player)

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

		// Process user input and recalculate view matrix
		if renderers.IsWindowFocused() {
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
		if renderers.Win.GetMouseButton(glfw.MouseButton1) == glfw.Press && mouseOne.Invoke() {
			core.TraceRay(player.LookDir(), player.CameraPos(), 16, func(v, _ mgl32.Vec3) (stop bool) {
				block := dim.GetBlockAt(core.NewVec3FromFloat(v))
				if block.Type != nil {
					block.Type = nil
					dim.SetBlockAt(block, core.NewVec3FromFloat(v))
					renderers.UpdateRequiredMeshes(dim, core.NewVec3FromFloat(v))
					return true
				} else {
					return false
				}
			})
		} else if renderers.Win.GetMouseButton(glfw.MouseButton1) == glfw.Release {
			mouseOne.Reset()
		}

		// Placing
		if renderers.Win.GetMouseButton(glfw.MouseButton2) == glfw.Press && mouseTwo.Invoke() {
			core.TraceRay(player.LookDir(), player.CameraPos(), 16, func(v, h mgl32.Vec3) (stop bool) {
				block := dim.GetBlockAt(core.NewVec3FromFloat(v))
				if block.Type != nil {
					placePos := core.NewVec3FromFloat(v)

					switch core.FaceFromSubvoxel(h) {
					case core.FaceTop:
						placePos = placePos.Add(core.Vec3{Y: 1})
					case core.FaceBottom:
						placePos = placePos.Add(core.Vec3{Y: -1})
					case core.FaceLeft:
						placePos = placePos.Add(core.Vec3{X: -1})
					case core.FaceRight:
						placePos = placePos.Add(core.Vec3{X: 1})
					case core.FaceFront:
						placePos = placePos.Add(core.Vec3{Z: 1})
					case core.FaceBack:
						placePos = placePos.Add(core.Vec3{Z: -1})
					}

					dim.SetBlockAt(core.Block{
						Type: blocks.Cobblestone,
					}, placePos)
					renderers.UpdateRequiredMeshes(dim, placePos)
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

		// Render gui
		gui.RenderGame()

		// Render all entities
		for _, v := range dim.Entities {
			core.EntityRegistry[v.EntityType].RenderType.RenderEntity(v, view)
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

				}
			default:
				break outer
			}
		}
	}
}
