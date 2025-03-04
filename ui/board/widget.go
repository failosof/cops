package chessboard

import (
	"image"
	"image/color"
	"log/slog"
	"slices"
	"sync"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget/material"
	"github.com/failosof/cops/ui/board/union"
	"github.com/failosof/cops/ui/board/util"
	"github.com/notnil/chess"
)

// todo: fix captures on the board
// todo: move out board state update
// todo: add inside coordinates
// todo: add each square coordinates
// todo: add pawn promotion
// todo: animations
// todo: remove core deps

type Widget struct {
	th *material.Theme

	config *Config

	redraw bool

	curBoardSize  union.Size
	prevBoardSize union.Size
	squareSize    union.Size
	hintSize      union.Size
	pointerSize   union.Size

	buttonPressed pointer.Buttons
	modifiersUsed key.Modifiers

	annoType    AnnoType
	drawingAnno Annotation
	annotations []*Annotation

	squareOrigins []union.Point

	pieceEventTargets []event.Filter

	coordsDrawingOp  op.CallOp
	boardDrawingOp   op.CallOp
	hintDrawingOp    op.CallOp
	squareDrawingOps []*op.CallOp

	dragID         pointer.ID
	draggingPos    union.Point
	selectedSquare chess.Square
	selectedPiece  chess.Piece

	flipped      bool
	game         *chess.Game
	curPosition  *chess.Position
	prevPosition *chess.Position
	promoteOn    chess.Square

	mu sync.Mutex
}

func NewWidget(th *material.Theme, config *Config) *Widget {
	w := Widget{
		th:                th,
		config:            config,
		pointerSize:       union.SizeFromInt(16), // assume for now
		squareOrigins:     make([]union.Point, 64),
		pieceEventTargets: make([]event.Filter, 64),
		squareDrawingOps:  make([]*op.CallOp, 64),
		selectedSquare:    chess.NoSquare,
		selectedPiece:     chess.NoPiece,
		annoType:          CircleAnno,
		game:              chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		promoteOn:         chess.NoSquare,
	}

	return &w
}

func (w *Widget) Layout(gtx layout.Context) layout.Dimensions {
	return CoordinatesStyle{
		Type:     w.config.Coordinates,
		Theme:    w.th,
		FontSize: 16,
		Flipped:  w.flipped,
		Board:    w.layout,
	}.Layout(gtx)
}

func (w *Widget) layout(gtx layout.Context) layout.Dimensions {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.curBoardSize = union.SizeFromMinPt(gtx.Constraints.Max)
	w.curPosition = w.game.Position()
	w.redraw = w.redraw || !w.curBoardSize.Eq(w.prevBoardSize) || w.PositionChanged()
	defer func() {
		w.redraw = false
		w.prevBoardSize = w.curBoardSize
		w.prevPosition = w.curPosition
	}()

	if w.redraw {
		w.squareSize = union.SizeFromFloat(w.curBoardSize.Float / 8)
		w.hintSize = union.SizeFromMinF32(w.squareSize.F32.Div(3))
		w.draggingPos = union.PointFromF32(w.draggingPos.F32)

		for square := chess.A1; square <= chess.H8; square++ {
			w.squareOrigins[square] = union.PointFromF32(util.SquareToPoint(square, w.squareSize.Float, w.flipped))
		}

		cache := new(op.Ops)
		boardMacro := op.Record(cache)
		factor := w.curBoardSize.F32.Div(w.config.BoardTexture.Size.Float)
		util.DrawImage(cache, w.config.BoardTexture.Image, image.Point{}, factor)
		w.boardDrawingOp = boardMacro.Stop()
	}

	w.boardDrawingOp.Add(gtx.Ops)

	defer clip.Rect(image.Rectangle{Max: w.curBoardSize.Pt}).Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, w)

	if w.config.ShowLastMove {
		lastMove := w.getLastMove()
		if lastMove != nil {
			w.markSquare(gtx, lastMove.S1(), w.config.Colors.LastMove)
			w.markSquare(gtx, lastMove.S2(), w.config.Colors.LastMove)
		}
	}

	if w.selectedSquare != chess.NoSquare && w.selectedPiece.Color() == w.curPosition.Turn() {
		w.markSquare(gtx, w.selectedSquare, w.config.Colors.Highlight)
		if w.config.ShowHints {
			for _, move := range w.curPosition.ValidMoves() {
				if move.S1() == w.selectedSquare {
					position := w.squareOrigins[move.S2()]
					if w.curPosition.Board().Piece(move.S2()) == chess.NoPiece {
						origin := position.F32.Add(w.squareSize.Half.F32).Sub(w.hintSize.Half.F32).Round()
						util.DrawEllipse(gtx.Ops, util.Rect(origin, w.hintSize.Pt), w.config.Colors.Hint)
					} else {
						rect := util.Rect(position.Pt, w.squareSize.Pt)
						util.DrawRectangle(gtx.Ops, rect, w.squareSize.Float/5, w.config.Colors.Hint)
					}
				}
			}
		}
	}

	w.drawPieces(gtx)

	for _, anno := range w.annotations {
		anno.Width = union.SizeFromFloat(w.squareSize.Float / 7)
		anno.Draw(gtx, w.squareOrigins, w.squareSize, w.redraw)
	}
	w.drawingAnno.Width = union.SizeFromFloat(w.squareSize.Float / 9)
	w.drawingAnno.Draw(gtx, w.squareOrigins, w.squareSize, w.drawingAnno.Type != NoAnno)

	for {
		ev, ok := gtx.Event(pointer.Filter{
			Target: w,
			Kinds:  pointer.Move | pointer.Press | pointer.Release | pointer.Drag,
		})
		if !ok {
			break
		}

		if e, ok := ev.(pointer.Event); ok {
			switch e.Kind {
			case pointer.Move:
				pointer.CursorPointer.Add(gtx.Ops)
			case pointer.Drag:
				if w.buttonPressed == pointer.ButtonSecondary {
					w.processSecondaryButtonDragging(gtx, e)
				}
			case pointer.Press:
				w.buttonPressed = e.Buttons
				w.modifiersUsed = e.Modifiers
				w.promoteOn = chess.NoSquare
				fallthrough
			default:
				if w.buttonPressed == pointer.ButtonPrimary {
					w.processPrimaryButtonClick(gtx, e)
				} else if w.buttonPressed == pointer.ButtonSecondary {
					w.processSecondaryButtonClick(gtx, e)
				}
			}
		}
	}

	for {
		ev, ok := gtx.Event(w.pieceEventTargets...)
		if !ok {
			break
		}

		if e, ok := ev.(pointer.Event); ok {
			switch e.Kind {
			case pointer.Move:
				pointer.CursorGrab.Add(gtx.Ops)
			case pointer.Drag:
				if w.buttonPressed == pointer.ButtonPrimary {
					w.processPrimaryButtonDragging(gtx, e)
				}
			case pointer.Release:
				w.processPrimaryButtonClick(gtx, e)
			}
		}
	}

	w.markSquare(gtx, w.promoteOn, w.config.Colors.Highlight)
	if w.promoteOn != chess.NoSquare {
		Promotion{
			Position:      w.squareOrigins[w.promoteOn],
			SquareSize:    w.squareSize,
			Color:         w.selectedPiece.Color(),
			Background:    w.config.Colors.Empty,
			PieceTextures: w.config.PieceTextures,
			Flipped:       w.flipped,
		}.Layout(gtx)
	}

	return layout.Dimensions{Size: w.curBoardSize.Pt}
}

func (w *Widget) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.game = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	w.curPosition = w.game.Position()
	w.redraw = true
}

func (w *Widget) MoveBackward() {
	w.mu.Lock()
	defer w.mu.Unlock()

	moves := w.game.Moves()
	if len(moves) > 0 {
		w.game = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
		for _, move := range moves[:len(moves)-1] {
			if err := w.game.Move(move); err != nil {
				slog.Warn("can't replay game", "move", move.String(), "err", err)
			}
		}
		w.curPosition = w.game.Position()
		w.redraw = true
	}
}

func (w *Widget) Flip() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.putSelectedPieceBack()
	w.unselectPiece()
	w.flipped = !w.flipped
	w.redraw = true
}

func (w *Widget) PositionChanged() bool {
	return w.prevPosition != nil && w.prevPosition.Hash() != w.curPosition.Hash()
}

func (w *Widget) Game() *chess.Game {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.game
}

func (w *Widget) drawPieces(gtx layout.Context) {
	if w.game == nil {
		return
	}

	if w.redraw {
		clear(w.squareDrawingOps)
		var wg sync.WaitGroup
		for square := chess.A1; square <= chess.H8; square++ {
			origin := w.squareOrigins[square]
			if piece := w.curPosition.Board().Piece(square); piece != chess.NoPiece {
				wg.Add(1)
				go func(square chess.Square, piece chess.Piece) {
					defer wg.Done()
					factor := w.squareSize.F32.Div(w.config.PieceTextures[piece].Size.Float)
					cache := new(op.Ops)
					squareMacro := op.Record(cache)
					util.DrawImage(cache, w.config.PieceTextures[piece].Image, origin.Pt, factor)
					ops := squareMacro.Stop()
					w.squareDrawingOps[square] = &ops
				}(square, piece)
			}
		}
		wg.Wait()
	}

	clear(w.pieceEventTargets)
	for square := chess.A1; square <= chess.H8; square++ {
		squareDrawingOp := w.squareDrawingOps[square]
		if squareDrawingOp != nil {
			origin := w.squareOrigins[square]
			pieceClip := clip.Rect(util.Rect(origin.Pt, w.squareSize.Pt)).Push(gtx.Ops)
			event.Op(gtx.Ops, square)
			pieceClip.Pop()
			w.pieceEventTargets = append(w.pieceEventTargets, pointer.Filter{
				Target: square,
				Kinds:  pointer.Move | pointer.Drag | pointer.Release,
			})

			if square != w.selectedSquare {
				squareDrawingOp.Add(gtx.Ops)
			}
		}
	}

	if w.selectedSquare != chess.NoSquare && w.promoteOn == chess.NoSquare {
		img := w.config.PieceTextures[w.selectedPiece].Image
		factor := w.squareSize.F32.Div(w.config.PieceTextures[w.selectedPiece].Size.Float)
		util.DrawImage(gtx.Ops, img, w.draggingPos.Pt, factor)
	}
}

func (w *Widget) processPrimaryButtonClick(gtx layout.Context, e pointer.Event) {
	hoveredSquare := util.PointToSquare(e.Position, w.squareSize.Float, w.flipped)
	if hoveredSquare == chess.NoSquare {
		return
	}
	hoveredPiece := w.game.Position().Board().Piece(hoveredSquare)

	defer gtx.Execute(op.InvalidateCmd{})

	switch e.Kind {
	case pointer.Press:
		clear(w.annotations)
		w.annotations = nil
		w.drawingAnno.Type = NoAnno

		if w.selectedPiece == chess.NoPiece || w.selectedPiece.Color() == hoveredPiece.Color() {
			if hoveredPiece != chess.NoPiece {
				w.selectPiece(gtx, e, hoveredPiece, hoveredSquare)
				return
			}
		}

		fallthrough
	case pointer.Release:
		if w.selectedSquare == hoveredSquare {
			w.putSelectedPieceBack()
			return
		}

		if w.selectedSquare != chess.NoSquare && w.selectedPiece != chess.NoPiece {
			move := w.selectedSquare.String() + hoveredSquare.String()
			if err := w.game.MoveStr(move); err != nil {
				w.putSelectedPieceBack()
			}

			//for _, validMove := range w.game.ValidMoves() {
			//	if strings.HasPrefix(validMove.String(), move) {
			//		if util.IsPromotionMove(hoveredSquare, w.selectedPiece) {
			//			w.promoteOn = hoveredSquare
			//			return
			//		}
			//
			//
			//
			//		break
			//	}
			//}

			//if hoveredPiece != chess.NoPiece {
			//	w.selectPiece(gtx, e, hoveredPiece, hoveredSquare)
			//	return
			//}
		}

		fallthrough
	default:
		w.unselectPiece()
		w.buttonPressed = 0
		w.modifiersUsed = 0
	}
}

func (w *Widget) processSecondaryButtonClick(gtx layout.Context, e pointer.Event) {
	hoveredSquare := util.PointToSquare(e.Position, w.squareSize.Float, w.flipped)
	defer gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(time.Second / 30)})

	switch e.Kind {
	case pointer.Press:
		if hoveredSquare != chess.NoSquare {
			w.drawingAnno = Annotation{
				Type:  w.annoType,
				Start: hoveredSquare,
				Color: w.selectAnnotationColor(),
				Width: union.SizeFromFloat(w.squareSize.Float / 9),
			}
			w.dragID = e.PointerID
		}
	case pointer.Release:
		w.drawingAnno.Width = union.SizeFromFloat(w.squareSize.Float / 7)
		w.drawingAnno.Color = w.selectAnnotationColor()
		if hoveredSquare != chess.NoSquare {
			w.drawingAnno.End = hoveredSquare
		}

		i := slices.IndexFunc(w.annotations, func(annotation *Annotation) bool {
			if w.drawingAnno.Type != annotation.Type {
				return false
			}
			if w.drawingAnno.Type == ArrowAnno {
				return annotation.Start == w.drawingAnno.Start && annotation.End == w.drawingAnno.End
			} else {
				return annotation.Start == w.drawingAnno.Start
			}
		})

		anno := w.drawingAnno.Copy()
		if i > -1 {
			if w.annotations[i].Equal(&w.drawingAnno) {
				w.annotations = slices.Delete(w.annotations, i, i+1)
			} else {
				w.annotations[i] = &anno
			}
		} else {
			w.annotations = append(w.annotations, &anno)
		}

		w.drawingAnno = Annotation{}
		w.dragID = 0
		w.buttonPressed = 0
		w.modifiersUsed = 0
	}
}

func (w *Widget) processPrimaryButtonDragging(gtx layout.Context, e pointer.Event) {
	if w.dragID == e.PointerID && w.selectedSquare != chess.NoSquare {
		pointer.CursorGrabbing.Add(gtx.Ops)
		w.dragTo(gtx, e.Position)
	}
}

func (w *Widget) processSecondaryButtonDragging(gtx layout.Context, e pointer.Event) {
	if w.drawingAnno.Type != NoAnno {
		hoveredSquare := util.PointToSquare(e.Position, w.squareSize.Float, w.flipped)
		if hoveredSquare != chess.NoSquare {
			if w.dragID == e.PointerID {
				w.drawingAnno.End = hoveredSquare
				if w.drawingAnno.Start == w.drawingAnno.End {
					w.drawingAnno.Type = w.annoType
				} else {
					w.drawingAnno.Type = ArrowAnno
				}
				gtx.Execute(op.InvalidateCmd{})
			}
		}
	}
}

func (w *Widget) selectPiece(gtx layout.Context, e pointer.Event, piece chess.Piece, square chess.Square) {
	if piece != chess.NoPiece && square != chess.NoSquare {
		pointer.CursorGrabbing.Add(gtx.Ops)
		w.dragID = e.PointerID
		w.selectedPiece = piece
		w.selectedSquare = square
		w.dragTo(gtx, e.Position)
	}
}

func (w *Widget) dragTo(gtx layout.Context, pos f32.Point) {
	w.draggingPos = union.PointFromF32(pos.Add(w.pointerSize.Half.F32).Sub(w.squareSize.Half.F32))
	gtx.Execute(pointer.GrabCmd{
		Tag: w.selectedSquare,
		ID:  w.dragID,
	})
}

func (w *Widget) putSelectedPieceBack() {
	if w.selectedSquare != chess.NoSquare {
		w.draggingPos = w.squareOrigins[w.selectedSquare]
	}
}

func (w *Widget) unselectPiece() {
	if w.selectedSquare != chess.NoSquare {
		w.draggingPos = w.squareOrigins[w.selectedSquare]
	}

	w.promoteOn = chess.NoSquare
	w.selectedSquare = chess.NoSquare
	w.selectedPiece = chess.NoPiece
	w.dragID = 0
}

func (w *Widget) getLastMove() (m *chess.Move) {
	moves := w.game.Moves()
	if len(moves) > 0 {
		m = moves[len(moves)-1]
	}
	return
}

func (w *Widget) markSquare(gtx layout.Context, square chess.Square, color color.NRGBA) {
	if square != chess.NoSquare {
		origin := w.squareOrigins[square]
		util.DrawPane(gtx.Ops, util.Rect(origin.Pt, w.squareSize.Pt), color)
	}
}

func (w *Widget) selectAnnotationColor() color.NRGBA {
	if w.modifiersUsed == 0 {
		return w.config.Colors.Primary
	}

	if w.modifiersUsed&key.ModAlt == key.ModAlt {
		return w.config.Colors.Warning
	} else if w.modifiersUsed&key.ModShift == key.ModShift {
		return w.config.Colors.Info
	} else if w.modifiersUsed&key.ModCtrl == key.ModCtrl {
		return w.config.Colors.Danger
	} else {
		return w.config.Colors.Primary
	}
}

func (w *Widget) SetGame(game *chess.Game) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.game = game
}
