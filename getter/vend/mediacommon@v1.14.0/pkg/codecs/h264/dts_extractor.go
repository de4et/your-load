package h264

import (
	"bytes"
	"fmt"
	"time"
)

// DTSExtractor allows to extract DTS from PTS.
//
// Deprecated: replaced by DTSExtractor2.
type DTSExtractor struct {
	sps             []byte
	spsp            *SPS
	prevDTSFilled   bool
	prevDTS         time.Duration
	expectedPOC     uint32
	reorderedFrames int
	pauseDTS        int
	pocIncrement    int
}

// NewDTSExtractor allocates a DTSExtractor.
//
// Deprecated: replaced by NewDTSExtractor2.
func NewDTSExtractor() *DTSExtractor {
	return &DTSExtractor{
		pocIncrement: 2,
	}
}

func (d *DTSExtractor) extractInner(au [][]byte, pts time.Duration) (time.Duration, bool, error) {
	var idr []byte
	var nonIDR []byte
	// a value of 00 indicates that the content of the NAL unit is not
	// used to reconstruct reference pictures for inter picture
	// prediction.  Such NAL units can be discarded without risking
	// the integrity of the reference pictures.  Values greater than
	// 00 indicate that the decoding of the NAL unit is required to
	// maintain the integrity of the reference pictures.
	nonZeroNalRefIDFound := false

	for _, nalu := range au {
		typ := NALUType(nalu[0] & 0x1F)
		nonZeroNalRefIDFound = nonZeroNalRefIDFound || ((nalu[0] & 0x60) > 0)
		switch typ {
		case NALUTypeSPS:
			if !bytes.Equal(d.sps, nalu) {
				var spsp SPS
				err := spsp.Unmarshal(nalu)
				if err != nil {
					return 0, false, fmt.Errorf("invalid SPS: %w", err)
				}
				d.sps = nalu
				d.spsp = &spsp

				// reset state
				d.reorderedFrames = 0
				d.pocIncrement = 2
			}

		case NALUTypeIDR:
			idr = nalu

		case NALUTypeNonIDR:
			nonIDR = nalu
		}
	}

	if d.spsp == nil {
		return 0, false, fmt.Errorf("SPS not received yet")
	}

	if d.spsp.PicOrderCntType == 2 || !d.spsp.FrameMbsOnlyFlag {
		return pts, false, nil
	}

	if d.spsp.PicOrderCntType == 1 {
		return 0, false, fmt.Errorf("pic_order_cnt_type = 1 is not supported yet")
	}

	// Implicit processing of PicOrderCountType 0
	switch {
	case idr != nil:
		d.pauseDTS = 0

		var err error
		d.expectedPOC, err = getPictureOrderCount(idr, d.spsp, true)
		if err != nil {
			return 0, false, err
		}

		if !d.prevDTSFilled || d.reorderedFrames == 0 {
			return pts, false, nil
		}

		return d.prevDTS + (pts-d.prevDTS)/time.Duration(d.reorderedFrames+1), false, nil

	case nonIDR != nil:
		d.expectedPOC += uint32(d.pocIncrement)
		d.expectedPOC &= ((1 << (d.spsp.Log2MaxPicOrderCntLsbMinus4 + 4)) - 1)

		if d.pauseDTS > 0 {
			d.pauseDTS--
			return d.prevDTS + 1*time.Millisecond, true, nil
		}

		poc, err := getPictureOrderCount(nonIDR, d.spsp, false)
		if err != nil {
			return 0, false, err
		}

		if d.pocIncrement == 2 && (poc%2) != 0 {
			d.pocIncrement = 1
			d.expectedPOC /= 2
		}

		pocDiff := int(getPictureOrderCountDiff(poc, d.expectedPOC, d.spsp)) / d.pocIncrement
		limit := -(d.reorderedFrames + 1)

		// this happens when there are B-frames immediately following an IDR frame
		if pocDiff < limit {
			increase := limit - pocDiff
			if (d.reorderedFrames + increase) > maxReorderedFrames {
				return 0, false, fmt.Errorf("too many reordered frames (%d)", d.reorderedFrames+increase)
			}

			d.reorderedFrames += increase
			d.pauseDTS = increase
			return d.prevDTS + 1*time.Millisecond, true, nil
		}

		if pocDiff == limit {
			return pts, false, nil
		}

		if pocDiff > d.reorderedFrames {
			increase := pocDiff - d.reorderedFrames
			if (d.reorderedFrames + increase) > maxReorderedFrames {
				return 0, false, fmt.Errorf("too many reordered frames (%d)", d.reorderedFrames+increase)
			}

			d.reorderedFrames += increase
			d.pauseDTS = increase - 1
			return d.prevDTS + 1*time.Millisecond, false, nil
		}

		return d.prevDTS + (pts-d.prevDTS)/time.Duration(pocDiff+d.reorderedFrames+1), false, nil

	case !nonZeroNalRefIDFound:
		return d.prevDTS, false, nil

	default:
		return 0, false, fmt.Errorf("access unit doesn't contain an IDR or non-IDR NALU")
	}
}

// Extract extracts the DTS of an access unit.
func (d *DTSExtractor) Extract(au [][]byte, pts time.Duration) (time.Duration, error) {
	dts, skipChecks, err := d.extractInner(au, pts)
	if err != nil {
		return 0, err
	}

	if !skipChecks && dts > pts {
		return 0, fmt.Errorf("DTS is greater than PTS")
	}

	if d.prevDTSFilled && dts < d.prevDTS {
		return 0, fmt.Errorf("DTS is not monotonically increasing, was %v, now is %v",
			d.prevDTS, dts)
	}

	d.prevDTS = dts
	d.prevDTSFilled = true

	return dts, err
}
