// Copyright ©2023 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
 * Copyright (c) 2014 Kurt Jung (Gmail: kurt.w.jung)
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package tinypdf

// SVGBasicWrite renders the paths encoded in the basic SVG image specified by
// sb. The scale value is used to convert the coordinates in the path to the
// unit of measure specified in New(). If scale is 0, SVGBasicWrite automatically adapts the SVG document
// to the PDF document unit. The current position (as set with a call
// to SetXY()) is used as the origin of the image. The current line cap style
// (as set with SetLineCapStyle()), line width (as set with SetLineWidth()),
// and draw color (as set with SetDrawColor()) are used in drawing the image
// paths.
func (f *DocPDF) SVGBasicWrite(sb *SVGBasicType, scale float64) {
	originX, originY := f.GetXY()
	var x, y, newX, newY float64
	var cx0, cy0, cx1, cy1 float64
	var path []SVGBasicSegmentType
	var seg SVGBasicSegmentType
	var startX, startY float64
	if scale == 0.0 {
		scale = 1.0 / f.k
	}
	sval := func(origin float64, arg int) float64 {
		return origin + scale*seg.Arg[arg]
	}
	xval := func(arg int) float64 {
		return sval(originX, arg)
	}
	yval := func(arg int) float64 {
		return sval(originY, arg)
	}
	val := func(arg int) (float64, float64) {
		return xval(arg), yval(arg + 1)
	}
	for j := 0; j < len(sb.Segments) && f.Ok(); j++ {
		path = sb.Segments[j]
		for k := 0; k < len(path) && f.Ok(); k++ {
			seg = path[k]
			switch seg.Cmd {
			case 'M':
				x, y = val(0)
				startX, startY = x, y
				f.SetXY(x, y)
			case 'L':
				newX, newY = val(0)
				f.Line(x, y, newX, newY)
				x, y = newX, newY
			case 'C':
				cx0, cy0 = val(0)
				cx1, cy1 = val(2)
				newX, newY = val(4)
				f.CurveCubic(x, y, cx0, cy0, newX, newY, cx1, cy1, "D")
				x, y = newX, newY
			case 'Q':
				cx0, cy0 = val(0)
				newX, newY = val(2)
				f.Curve(x, y, cx0, cy0, newX, newY, "D")
				x, y = newX, newY
			case 'H':
				newX = xval(0)
				f.Line(x, y, newX, y)
				x = newX
			case 'V':
				newY = yval(0)
				f.Line(x, y, x, newY)
				y = newY
			case 'Z':
				f.Line(x, y, startX, startY)
				x, y = startX, startY
			default:
				f.SetErrorf("Unexpected path command '%c'", seg.Cmd)
			}
		}
	}
}

// SVGBasicDraw renders the paths in the provided SVGBasicType, but each SVG shape is written
// as a path that can be filled.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or
// "FD" for outlined and filled. An empty string will be replaced with
// "D". Drawing uses the current draw color and line width centered on
// the ellipse's perimeter. Filling uses the current fill color.
func (f *DocPDF) SVGBasicDraw(sb *SVGBasicType, scale float64, styleStr string) {
	originX, originY := f.GetXY()
	var newX, newY float64
	var cx0, cy0, cx1, cy1 float64
	var path []SVGBasicSegmentType
	var seg SVGBasicSegmentType
	var startX, startY float64
	if scale == 0.0 {
		scale = 1.0 / f.k
	}
	sval := func(origin float64, arg int) float64 {
		return origin + scale*seg.Arg[arg]
	}
	xval := func(arg int) float64 {
		return sval(originX, arg)
	}
	yval := func(arg int) float64 {
		return sval(originY, arg)
	}
	val := func(arg int) (float64, float64) {
		return xval(arg), yval(arg + 1)
	}
	for j := 0; j < len(sb.Segments) && f.Ok(); j++ {
		path = sb.Segments[j]
		for k := 0; k < len(path) && f.Ok(); k++ {
			seg = path[k]
			switch seg.Cmd {
			case 'M':
				startX, startY = val(0)
				f.MoveTo(startX, startY)
			case 'L':
				newX, newY = val(0)
				f.LineTo(newX, newY)
			case 'C':
				cx0, cy0 = val(0)
				cx1, cy1 = val(2)
				newX, newY = val(4)
				f.CurveBezierCubicTo(cx0, cy0, cx1, cy1, newX, newY)
			case 'Q':
				cx0, cy0 = val(0)
				newX, newY = val(2)
				f.CurveTo(cx0, cy0, newX, newY)
			case 'H':
				newX = xval(0)
				f.LineTo(newX, f.GetY())
			case 'V':
				newY = yval(0)
				f.LineTo(f.GetX(), newY)
			case 'Z':
				f.ClosePath()
				f.DrawPath(styleStr)
			default:
				f.SetErrorf("Unexpected path command '%c'", seg.Cmd)
			}
		}
	}
}
