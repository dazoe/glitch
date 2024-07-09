package glitch

type meshDraw struct {
	mesh *Mesh
	matrix glMat4
	mask RGBA
	material Material
	translucent bool
}

// For batching multiple sprites into one
type DrawBatch struct {
	draws []meshDraw

	boundsSet bool
	bounds Box
}

func NewDrawBatch() *DrawBatch {
	return &DrawBatch{
		draws: make([]meshDraw, 0),
	}
}

// func (b *DrawBatch) Buffer(pass *RenderPass) *DrawBatch {
// 	return &DrawBatch{
// 		mesh: b.mesh.Buffer(pass, b.material, b.Translucent),
// 		material: b.material,
// 		Translucent: b.Translucent,
// 	}
// }

func (b *DrawBatch) Add(mesh *Mesh, matrix glMat4, mask RGBA, material Material, translucent bool) {
	b.draws = append(b.draws, meshDraw{
		mesh: mesh,
		matrix: matrix,
		mask: mask,
		material: material,
		translucent: translucent,
	})

	newBounds := mesh.Bounds().Apply(matrix)
	// TODO: Does this improve performance?
	// if matrix != glMat4Ident {
	// 	newBounds = newBounds.Apply(matrix)
	// }
	if b.boundsSet {
		b.bounds = b.bounds.Union(newBounds)
	} else {
		b.boundsSet = true
		b.bounds = newBounds
	}
}

func (b *DrawBatch) Clear() {
	b.draws = b.draws[:0]
	b.boundsSet = false
	b.bounds = Box{}
}

func (b *DrawBatch) Draw(target BatchTarget, matrix Mat4) {
	for i := range b.draws {
		mat := matrix.gl()
		mat.Mul(&b.draws[i].matrix)
		target.Add(b.draws[i].mesh, mat, b.draws[i].mask, b.draws[i].material, b.draws[i].translucent)
	}
	// target.Add(b.mesh, matrix.gl(), RGBA{1.0, 1.0, 1.0, 1.0}, b.material, b.Translucent)
	// b.DrawColorMask(target, matrix, White)
}

func (b *DrawBatch) DrawColorMask(target BatchTarget, matrix Mat4, color RGBA) {
	for i := range b.draws {
		mat := matrix.gl()
		mat.Mul(&b.draws[i].matrix)

		mask := b.draws[i].mask.Mult(color)
		target.Add(b.draws[i].mesh, mat, mask, b.draws[i].material, b.draws[i].translucent)
	}

	// target.Add(b.mesh, matrix.gl(), color, b.material, b.Translucent)
	// for i := range b.draws {
	// 	target.Add(b.draws[i].mesh, b.draws[i].matrix, b.draws[i].color, b.draws[i].material, b.draws[i].translucent)
	// }
}

func (b *DrawBatch) RectDraw(target BatchTarget, bounds Rect) {
	batchBounds := b.Bounds().Rect()
	matrix := Mat4Ident
	matrix.Scale(bounds.W() / batchBounds.W(), bounds.H() / batchBounds.H(), 1).Translate(bounds.W()/2 + bounds.Min.X, bounds.H()/2 + bounds.Min.Y, 0)

	b.Draw(target, matrix)

	// b.RectDrawColorMask(target, bounds, RGBA{1, 1, 1, 1})
}

// TODO: Generalize this rectdraw logic. Copy paseted from Sprite
func (b *DrawBatch) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
	batchBounds := b.Bounds().Rect()
	matrix := Mat4Ident
	matrix.Scale(bounds.W() / batchBounds.W(), bounds.H() / batchBounds.H(), 1).Translate(bounds.W()/2 + bounds.Min.X, bounds.H()/2 + bounds.Min.Y, 0)

	b.DrawColorMask(target, matrix, mask)

	// // pass.SetTexture(0, s.texture)
	// // pass.Add(s.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, s.material)

	// batchBounds := b.Bounds().Rect()
	// matrix := Mat4Ident
	// matrix.Scale(bounds.W() / batchBounds.W(), bounds.H() / batchBounds.H(), 1).Translate(bounds.W()/2 + bounds.Min[0], bounds.H()/2 + bounds.Min[1], 0)
	// target.Add(b.mesh, matrix.gl(), mask, b.material, false)
}

func (b *DrawBatch) Bounds() Box {
	return b.bounds
}

// type Batcher struct {
// 	shader *Shader
// 	lastBuffer *VertexBuffer
// 	target Target
// }

// func NewBatcher() *Batcher {
// 	return &Batcher{} // TODO: Default case for shader?
// }

// func (b *Batcher) SetShader(shader *Shader) {
// 	b.Flush() // TODO: You technically only need to do this if it will change the uniform

// 	b.shader = shader
// }

// func (b *Batcher) SetUniform(name string, val any) {
// 	b.Flush() // TODO: You technically only need to do this if it will change the uniform

// 	b.shader.SetUniform(name, val)
// }

// func (b *Batcher) Clear() {

// }

// func (b *Batcher) Add(filler GeometryFiller, mat glMat4, mask RGBA, material Material, translucent bool) {
// 	if filler == nil { return } // Skip nil meshes

// 	buffer := filler.GetBuffer()
// 	if buffer != nil {
// 		b.drawCall(buffer, mat)
// 		return
// 	}

// 	// Note: Captured in shader.pool
// 	// 1. If you switch materials, then draw the last one
// 	// 2. If you fill up then draw the last one
// 	state := BufferState{material, BlendModeNormal} // TODO: blendmode and track full state some better way
// 	vertexBuffer := filler.Fill(b.shader.pool, mat, mask, state)

// 	// If vertexBuffer has changed then we want to draw the last one
// 	if b.lastBuffer != nil && vertexBuffer != b.lastBuffer {
// 		b.drawCall(b.lastBuffer, glMat4Ident)
// 	}

// 	b.lastBuffer = vertexBuffer
// }

// // Draws the current buffer and progress the shader pool to the next available
// func (b *Batcher) Flush() {
// 	if b.lastBuffer == nil { return }

// 	b.drawCall(b.lastBuffer, glMat4Ident)
// 	b.lastBuffer = nil
// 	b.shader.pool.gotoNextClean()
// }

// // Executes a drawcall with ...
// func (b *Batcher) drawCall(buffer *VertexBuffer, mat glMat4) {
// 	if b.target != nil {
// 		b.target.Bind()
// 	}

// 	// TODO: Set all uniforms
// 	// 1. camera
// 	// 2. materials

// 	b.shader.Bind() // TODO: global State cache

// 	// TODO: rewrite how buffer state works for immediate mode case
// 	buffer.state.Bind(b.shader)

// 	// TOOD: Maybe pass this into VertexBuffer.Draw() func
// 	ok := b.shader.SetUniform("model", mat)
// 	if !ok {
// 		panic("Error setting model uniform - all shaders must have 'model' uniform")
// 	}

// 	buffer.Draw()
// }

