package tinypdf

import (
	"github.com/cdvelop/tinypdf/fontManager"
	"path/filepath"
)

func isAbsolutePath(p string) bool {
	return filepath.IsAbs(p)
}

func (f *TinyPDF) generateCIDFontMap(font *fontManager.FontDefType, LastRune int) {
	rangeID := 0
	cidArray := make(map[int]*untypedKeyMap)
	cidArrayKeys := make([]int, 0)
	prevCid := -2
	prevWidth := -1
	interval := false
	startCid := 1
	cwLen := LastRune + 1

	// for each character
	for cid := startCid; cid < cwLen; cid++ {
		if font.Cw[cid] == 0x00 {
			continue
		}
		width := font.Cw[cid]
		if width == 65535 {
			width = 0
		}
		if numb, OK := font.UsedRunes[cid]; cid > 255 && (!OK || numb == 0) {
			continue
		}

		if cid == prevCid+1 {
			if width == prevWidth {

				if width == cidArray[rangeID].get(0) {
					cidArray[rangeID].put(nil, width)
				} else {
					cidArray[rangeID].pop()
					rangeID = prevCid
					r := untypedKeyMap{
						valueSet: make([]int, 0),
						keySet:   make([]interface{}, 0),
					}
					cidArray[rangeID] = &r
					cidArrayKeys = append(cidArrayKeys, rangeID)
					cidArray[rangeID].put(nil, prevWidth)
					cidArray[rangeID].put(nil, width)
				}
				interval = true
				cidArray[rangeID].put("interval", 1)
			} else {
				if interval {
					// new range
					rangeID = cid
					r := untypedKeyMap{
						valueSet: make([]int, 0),
						keySet:   make([]interface{}, 0),
					}
					cidArray[rangeID] = &r
					cidArrayKeys = append(cidArrayKeys, rangeID)
					cidArray[rangeID].put(nil, width)
				} else {
					cidArray[rangeID].put(nil, width)
				}
				interval = false
			}
		} else {
			rangeID = cid
			r := untypedKeyMap{
				valueSet: make([]int, 0),
				keySet:   make([]interface{}, 0),
			}
			cidArray[rangeID] = &r
			cidArrayKeys = append(cidArrayKeys, rangeID)
			cidArray[rangeID].put(nil, width)
			interval = false
		}
		prevCid = cid
		prevWidth = width

	}
	previousKey := -1
	nextKey := -1
	isInterval := false
	for g := 0; g < len(cidArrayKeys); {
		key := cidArrayKeys[g]
		ws := *cidArray[key]
		cws := len(ws.keySet)
		if (key == nextKey) && (!isInterval) && (ws.getIndex("interval") < 0 || cws < 4) {
			if cidArray[key].getIndex("interval") >= 0 {
				cidArray[key].delete("interval")
			}
			cidArray[previousKey] = arrayMerge(cidArray[previousKey], cidArray[key])
			cidArrayKeys = remove(cidArrayKeys, key)
		} else {
			g++
			previousKey = key
		}
		nextKey = key + cws
		// ui := ws.getIndex("interval")
		// ui = ui + 1
		if ws.getIndex("interval") >= 0 {
			if cws > 3 {
				isInterval = true
			} else {
				isInterval = false
			}
			cidArray[key].delete("interval")
			nextKey--
		} else {
			isInterval = false
		}
	}
	var w fmtBuffer
	for _, k := range cidArrayKeys {
		ws := cidArray[k]
		if len(arrayCountValues(ws.valueSet)) == 1 {
			w.printf(" %d %d %d", k, k+len(ws.valueSet)-1, ws.get(0))
		} else {
			w.printf(" %d [ %s ]\n", k, implode(" ", ws.valueSet))
		}
	}
	f.out("/W [" + w.String() + " ]")
}
