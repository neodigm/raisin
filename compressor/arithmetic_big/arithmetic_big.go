package arithmetic_big

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"sort"
)

type sortBytes []byte

const prec = 200

func Compress(input []byte) []byte {
	symFreqs := make(map[byte]int)
	total := len(input)
	for _, c := range input {
		symFreqs[c]++
	}
	symFreqsWhole := make(map[byte]*big.Float, len(symFreqs))
	for c, freq := range symFreqs {
		symFreqsWhole[c] = new(big.Float).SetPrec(prec).Quo(big.NewFloat(float64(freq)), big.NewFloat(float64(total)))
	}
	// symFreqsWhole = map[byte]float64{'3': 0.4, '2': 0.5, '1': 0.05, '0': 0.05}
	keys := make(sortBytes, 0)
	for k := range symFreqsWhole {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	fmt.Printf("-------------\n")
	for i := len(keys) - 1; i >= 0; i-- {
		fmt.Printf("%s - %f\n", string(keys[i]), symFreqsWhole[keys[i]])
	}
	fmt.Printf("-------------\n")
	top, bot := encode(keys, symFreqsWhole, input)
	fmt.Println(bot, "-", top, string(input))
	binaryLocation := getRootBinaryPosition(top, bot)
	fmt.Println(binaryLocation)
	return []byte("compress")
}

func encode(keys []byte, freqs map[byte]*big.Float, input []byte) (top *big.Float, bottom *big.Float) {
	if len(input) == 0 {
		return nil, nil
	}

	bottom, top = new(big.Float).SetPrec(prec), new(big.Float).SetPrec(prec)

	// Pop first byte off the input
	encodeByte := input[0]
	input = input[1:]

	sec := getSection(keys, freqs, encodeByte)
	for i := 0; i < sec; i++ {
		bottom.Add(bottom, freqs[keys[i]])
	}
	top.Add(bottom, freqs[keys[sec]])
	// fmt.Println("before", bottom, "-", top, string(encodeByte))

	// fmt.Println(getRootBinaryPosition(top, bottom))

	nextTop, nextBottom := encode(keys, freqs, input)
	// fmt.Println("next after", nextBottom, "-", nextTop)
	size := new(big.Float).SetPrec(prec)
	if nextBottom != nil && nextTop != nil {
		size.Sub(nextTop, nextBottom)
	}
	if nextBottom != nil {
		bottom.Add(bottom, new(big.Float).SetPrec(prec).Mul(freqs[keys[sec]], nextBottom))
	}
	if nextTop != nil {
		top.Add(bottom, new(big.Float).SetPrec(prec).Mul(freqs[keys[sec]], size))
	}

	return top, bottom
}

func getRootBinaryPosition(targetTop *big.Float, targetBot *big.Float) string {
	return getBinaryPosition(targetTop, targetBot, big.NewFloat(1), new(big.Float).SetPrec(prec))
}

func getBinaryPosition(targetTop *big.Float, targetBot *big.Float, top *big.Float, bot *big.Float) string {
	if targetTop.Cmp(top) >= 0 && targetBot.Cmp(bot) <= 0 {
		return ""
	}
	// fmt.Println("Current", bot, top, "Target", targetBot, targetTop)
	// fmt.Println("%d %d", targetTop.Cmp(top), targetBot.Cmp(bot))
	diff := new(big.Float).SetPrec(prec).Sub(top, bot)
	// fmt.Printf("%4g-%4g %4g-%4g\n", top, bot, targetTop, targetBot)
	targetDiff := new(big.Float).SetPrec(prec).Sub(targetTop, targetBot)
	targetHalf := new(big.Float).SetPrec(prec).Quo(targetDiff, big.NewFloat(2))
	targetHalfway := new(big.Float).SetPrec(prec).Sub(targetTop, targetHalf)
	halfwayPoint := new(big.Float).SetPrec(prec).Sub(top, new(big.Float).SetPrec(prec).Quo(diff, big.NewFloat(2)))
	// fmt.Printf("%4g vs. %4g\n", halfwayPoint, targetHalfway)
	if halfwayPoint.Cmp(targetHalfway) == -1 {
		// Target range is above halfway point
		// fmt.Println("Above")
		return "1" + getBinaryPosition(targetTop, targetBot, top, halfwayPoint)
	} else {
		// Target range is below halfway point
		// fmt.Println("Below")
		return "0" + getBinaryPosition(targetTop, targetBot, halfwayPoint, bot)
	}
}

func getSection(keys []byte, freqs map[byte]*big.Float, input byte) int {
	for i, key := range keys {
		if key == input {
			return i
		}
	}
	return -1
}

func Decompress(input []byte) []byte {
	return []byte("decompress")
}

func (s sortBytes) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortBytes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortBytes) Len() int {
	return len(s)
}

type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) io.WriteCloser {
	z := new(Writer)
	z.w = w
	return z
}

func (writer *Writer) Write(data []byte) (n int, err error) {
	compressed := Compress(data)
	writer.w.Write(compressed)
	return len(compressed), nil
}

func (writer *Writer) Close() error {
	return nil
}

type Reader struct {
	r            io.Reader
	compressed   []byte
	decompressed []byte
	pos          int
}

func NewReader(r io.Reader) io.Reader {
	z := new(Reader)
	z.r = r
	return z
}

func (r *Reader) Read(content []byte) (n int, err error) {
	if r.decompressed == nil {
		r.compressed, err = ioutil.ReadAll(r.r)
		if err != nil {
			return 0, err
		}
		r.decompressed = Decompress(r.compressed)
	}
	bytesToWriteOut := len(r.decompressed[r.pos:])
	if len(content) < bytesToWriteOut {
		bytesToWriteOut = len(content)
	}
	for i := 0; i < bytesToWriteOut; i++ {
		content[i] = r.decompressed[r.pos:][i]
	}
	if len(r.decompressed[r.pos:]) <= len(content) {
		err = io.EOF
	} else {
		r.pos += len(content)
	}
	return bytesToWriteOut, err
}

func (r *Reader) Close() error {
	return nil
}
