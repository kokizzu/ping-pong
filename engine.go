package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"
)

var engineLogger = log.New(os.Stdout, "[ENGINE] ", 0)

type canvasEngine struct {
	game Game

	ballX float64
	ballY float64
	p1X   float64
	p1Y   float64
	p2X   float64
	p2Y   float64

	p1YVelocity   float64
	p2YVelocity   float64
	ballXVelocity float64
	ballYVelocity float64

	// TODO Acceleration: Velocity increase per time

	fps float64
	tps float64
}

const (
	// Frames per second
	_fps = 50.0
	// Ticks per second
	_tps = 1000.0 * (1.0 / _fps)
)

func newCanvasEngine(g Game) *canvasEngine {
	e := &canvasEngine{}

	e.game = g

	e.fps = _fps
	e.tps = _tps

	return e
}

func (e *canvasEngine) bootstrap() *canvasEngine {
	engineLogger.Println("bootstrap ...")

	// Real random
	rand.Seed(time.Now().UnixNano())

	time.Sleep(750 * time.Millisecond)

	e.resetBall().resetPlayers().randomBall().log()

	return e
}

func (e *canvasEngine) writeFrames(gameCtx context.Context, frames chan []byte) {
	go func() {
		// tps ticks or millseconds for 1 frame, since: _tps * _fps = y
		ticker := time.NewTicker(_tps * time.Millisecond)
		ticks := 0

		for {
			select {
			case <-ticker.C:
				engineLogger.Println("next tick ...")

				if err := e.advanceGame(); err != nil {
					engineLogger.Println(err.Error())

					switch err {
					case ErrP1Win:
						e.game.p1Score += 1

					case ErrP2Win:
						e.game.p2Score += 1
					}

					// Reset
					e.bootstrap().writeFrames(gameCtx, frames)
					return
				}

				e.log()

				jsonFrame, err := e.jsonFrame()
				if err != nil {
					panic(err)
				}
				frames <- jsonFrame

				ticks++

				engineLogger.Printf("ticks %d", ticks)

			case <-gameCtx.Done():
				ticker.Stop()

				return
			}
		}
	}()
}

func (e *canvasEngine) jsonFrame() ([]byte, error) {
	return json.Marshal(e.mapFrame())
}

const (
	baseline                 = 0
	default_padding          = 15
	canvas_border_correction = 1

	default_ball_x_vel_ratio = 0.28
	min_ball_y_vel_ratio     = 0.1
	max_y_vel_ratio          = 0.25

	magic_p = 3
)

func (e *canvasEngine) ballDirP1() bool {
	return e.ballX <= e.game.width/2
}

func (e *canvasEngine) ballDirP2() bool {
	return e.ballX >= e.game.width/2
}

func (e *canvasEngine) log() *canvasEngine {
	jsonBytes, err := json.MarshalIndent(e.mapFrame(), "", "	")
	if err != nil {
		panic(err)
	}
	engineLogger.Printf("%s", jsonBytes)
	return e
}

// TODO Convert to int here already
func (e *canvasEngine) mapFrame() map[string]interface{} {
	return map[string]interface{}{
		"debug":      e.game.debug,
		"p1Score":    e.game.p1Score,
		"p2Score":    e.game.p2Score,
		"gameWidth":  e.game.width,
		"gameHeight": e.game.height,
		"p1Width":    e.game.p1.width,
		"p1Height":   e.game.p1.height,
		"p2Width":    e.game.p2.width,
		"p2Height":   e.game.p2.height,
		"ballWidth":  e.game.ball.width,
		"ballHeight": e.game.ball.height,

		// There are no half pixel
		"ballX":         int(e.ballX),
		"ballY":         int(e.ballY),
		"p1X":           int(e.p1X),
		"p1Y":           int(e.p1Y),
		"p2X":           int(e.p2X),
		"p2Y":           int(e.p2Y),
		"p1YVelocity":   int(e.p1YVelocity),
		"p2YVelocity":   int(e.p2YVelocity),
		"ballXVelocity": int(e.ballXVelocity),
		"ballYVelocity": int(e.ballYVelocity),
	}
}
