package abtest

import (
	"fmt"
	"regexp"
	"strings"
)

type Partitions interface {
	Init(string, int64) error
	String() string
	Clone() Partitions
	Contains(Partitions) bool
	Intersects(Partitions) bool
	Add(Partitions) Partitions
	Sub(Partitions) Partitions
	Array() []int64
}

type IntervalList struct {
	str       string
	max       int64
	intervals []*Interval
}

func (p *IntervalList) Init(str string, max int64) (err error) {
	p.str = ""
	p.max = 0
	p.intervals = nil

	if max <= 0 {
		return
	}

	p.max = max
	if len(str) == 0 {
		return
	}

	defer func() {
		if err != nil {
			p.str = ""
			p.intervals = nil
		}
	}()

	var min int64 = -1

	segments := strings.Split(str, ",")
	reRange := regexp.MustCompile("^[0-9]+-[0-9]+$")
	reNumber := regexp.MustCompile("^(0|[1-9][0-9]*)$")
	for _, segment := range segments {
		var n int
		var i1, i2 int64

		switch {
		case reRange.MatchString(segment):
			// range format: %d-%d
			n, err = fmt.Sscanf(segment, "%d-%d", &i1, &i2)
			if err != nil || n != 2 {
				err = fmt.Errorf("syntax err near %s", segment)
				return
			}

			if i1 <= min || i1 > i2 {
				err = fmt.Errorf("err near %s, partitions should be monotonic increasing", segment)
				return
			}
			if i2 >= max {
				err = fmt.Errorf("partition: %d should be less than partitionCount: %d", i2, max)
				return
			}
			min = i2

			p.intervals = appendInterval(p.intervals, i1, i2+1)
		case reNumber.MatchString(segment):
			// number format: %d
			n, err = fmt.Sscanf(segment, "%d", &i1)
			if err != nil || n != 1 {
				err = fmt.Errorf("syntax err near %s", segment)
				return
			}

			if i1 <= min {
				err = fmt.Errorf("err near %s, partitions should be monotonic increasing", segment)
				return
			}
			if i1 >= max {
				err = fmt.Errorf("partition: %d should be less than partitionCount: %d", i1, max)
				return
			}
			min = i1

			p.intervals = appendInterval(p.intervals, i1, i1+1)
		default:
			err = fmt.Errorf("syntax err near %s", segment)
			return
		}
	}

	p.formatStr()

	return
}

func (p *IntervalList) String() string {
	return p.str
}

func (p *IntervalList) Clone() Partitions {
	dst := new(IntervalList)
	dst.str = p.str
	dst.max = p.max

	dst.intervals = make([]*Interval, 0, len(p.intervals))
	for _, interval := range p.intervals {
		dst.intervals = append(dst.intervals, &Interval{
			Left:  interval.Left,
			Right: interval.Right,
		})
	}

	return dst
}

func (p *IntervalList) Contains(qq Partitions) bool {
	q := qq.(*IntervalList)
	if p.max != q.max {
		return false
	}

	pi, length := 0, len(p.intervals)
	for _, qInterval := range q.intervals {
		ql, qr := qInterval.Left, qInterval.Right
		var contains bool
		for pi < length {
			pInterval := p.intervals[pi]
			pl, pr := pInterval.Left, pInterval.Right
			if pr < ql {
				pi++
				continue
			}

			contains = (pl <= ql && pr >= qr)
			break
		}
		if !contains {
			return false
		}
	}

	return true
}

func (p *IntervalList) Intersects(qq Partitions) bool {
	q := qq.(*IntervalList)
	if p.max != q.max {
		return false
	}

	pi, qi, pLength, qLength := 0, 0, len(p.intervals), len(q.intervals)
	for pi < pLength && qi < qLength {
		pl, pr := p.intervals[pi].Left, p.intervals[pi].Right
		ql, qr := q.intervals[qi].Left, q.intervals[qi].Right
		switch {
		case pr <= ql:
			pi++
		case qr <= pl:
			qi++
		default:
			return true
		}
	}

	return false
}

func (p *IntervalList) Add(qq Partitions) (r Partitions) {
	q := qq.(*IntervalList)

	if p.max != q.max {
		return p.Clone()
	}

	intervals := make([]*Interval, 0)
	pi, qi, pLength, qLength := 0, 0, len(p.intervals), len(q.intervals)
	if pLength == 0 {
		return q.Clone()
	}
	if qLength == 0 {
		return p.Clone()
	}

	pl, pr := p.intervals[pi].Left, p.intervals[pi].Right
	ql, qr := q.intervals[qi].Left, q.intervals[qi].Right

loop:
	for {
		switch {
		case pr < ql:
			intervals = appendInterval(intervals, pl, pr)
			pi++
			if pi >= pLength {
				intervals = appendInterval(intervals, ql, qr)
				break loop
			}
			pl, pr = p.intervals[pi].Left, p.intervals[pi].Right
		case qr < pl:
			intervals = appendInterval(intervals, ql, qr)
			qi++
			if qi >= qLength {
				intervals = appendInterval(intervals, pl, pr)
				break loop
			}
			ql, qr = q.intervals[qi].Left, q.intervals[qi].Right
		default:
			l, r := pl, pr
			if ql < l {
				l = ql
			}
			if qr > r {
				r = qr
			}
			intervals = appendInterval(intervals, l, r)
			pi++
			qi++
			if pi >= pLength || qi >= qLength {
				break loop
			}

			pl, pr = p.intervals[pi].Left, p.intervals[pi].Right
			ql, qr = q.intervals[qi].Left, q.intervals[qi].Right
		}
	}

	for _, interval := range p.intervals[pi:] {
		intervals = appendInterval(intervals, interval.Left, interval.Right)
	}

	for _, interval := range q.intervals[qi:] {
		intervals = appendInterval(intervals, interval.Left, interval.Right)
	}

	result := &IntervalList{
		max:       p.max,
		intervals: intervals,
	}
	result.formatStr()

	return result
}

func (p *IntervalList) Sub(qq Partitions) (r Partitions) {
	q := qq.(*IntervalList)

	if p.max != q.max {
		return p.Clone()
	}

	intervals := make([]*Interval, 0)
	pi, qi, pLength, qLength := 0, 0, len(p.intervals), len(q.intervals)
	if pLength == 0 || qLength == 0 {
		return p.Clone()
	}

	pl, pr := p.intervals[pi].Left, p.intervals[pi].Right
	ql, qr := q.intervals[qi].Left, q.intervals[qi].Right

loop:
	for {
		switch {
		case pr <= ql:
			intervals = appendInterval(intervals, pl, pr)
			pi++
			if pi >= pLength {
				break loop
			}
			pl, pr = p.intervals[pi].Left, p.intervals[pi].Right
		case qr <= pl:
			qi++
			if qi >= qLength {
				intervals = appendInterval(intervals, pl, pr)
				pi++
				break loop
			}
			ql, qr = q.intervals[qi].Left, q.intervals[qi].Right
		case pl < ql:
			intervals = appendInterval(intervals, pl, ql)
			pl = ql
		case pr > qr:
			pl = qr
		default:
			pi++
			if pi >= pLength {
				break loop
			}
			pl, pr = p.intervals[pi].Left, p.intervals[pi].Right
		}
	}

	for _, interval := range p.intervals[pi:] {
		intervals = appendInterval(intervals, interval.Left, interval.Right)
	}

	result := &IntervalList{
		max:       p.max,
		intervals: intervals,
	}
	result.formatStr()

	return result
}

func (p *IntervalList) Array() []int64 {
	array := make([]int64, 0)
	for _, interval := range p.intervals {
		l, r := interval.Left, interval.Right
		for l < r {
			array = append(array, l)
			l++
		}
	}

	return array
}

func (p *IntervalList) formatStr() {
	ss := make([]string, 0, len(p.intervals))
	for _, interval := range p.intervals {
		ss = append(ss, interval.String())
	}
	p.str = strings.Join(ss, ",")
}

type Interval struct {
	Left  int64
	Right int64
}

func (i *Interval) String() string {
	if i.Left == i.Right-1 {
		return fmt.Sprintf("%d", i.Left)
	}

	return fmt.Sprintf("%d-%d", i.Left, i.Right-1)
}

func appendInterval(intervals []*Interval, l, r int64) (result []*Interval) {
	result = intervals
	length := len(intervals)

	if length > 0 && result[length-1].Right >= l {
		result[length-1].Right = r
		return
	}

	result = append(result, &Interval{Left: l, Right: r})
	return
}
