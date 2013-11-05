package monkey

import "testing"

var rt *Runtime
var cx *Context
var script1 *Script
var script2 *Script
var script3 *Script
var script4 *Script

func init() {
	rt = NewRuntime(8 * 1024 * 1024)
	cx = rt.NewContext()

	cx.Eval(`function add(a,b){
	   return a + b
    }`)

	cx.Eval(`function ooxx(i, j) {
        b = 0;
        for (a = 0; a < 10000; a ++) {
            b += 1 / ((i+j)*(i+j+1)/2 + i + 1)
        }
        return b;
    }`)

	cx.DefineFunction("add2", func(cx *Context, argv []*Value) *Value {
		var a, _ = argv[0].ToInt()
		var b, _ = argv[1].ToInt()
		return cx.Int(a + b)
	})

	script1 = cx.Compile("1 + 1", "script1", 0)

	script2 = cx.Compile("add(1,1)", "script2", 0)

	script3 = cx.Compile("add2(1,1)", "script3", 0)

	script4 = cx.Compile("ooxx(21233, 3452122)", "script4", 0)
}

func ooxx(i, j float64) float64 {
	var b = 0.0
	for a := 0; a < 10000; a++ {
		b += 1 / ((i+j)*(i+j+1)/2 + i + 1)
	}
	return b
}

func Test_Script1(t *testing.T) {
	v := script1.Execute()

	if v == nil {
		t.Fatal()
	}

	if v.IsInt() == false {
		t.Fatal()
	}

	i, ok := v.ToInt()

	if i != 2 || !ok {
		t.Fatal()
	}
}

func Test_Script2(t *testing.T) {
	v := script2.Execute()

	if v == nil {
		t.Fatal()
	}

	if v.IsInt() == false {
		t.Fatal()
	}

	i, ok := v.ToInt()

	if i != 2 || !ok {
		t.Fatal()
	}
}

func Test_Script3(t *testing.T) {
	v := script3.Execute()

	if v == nil {
		t.Fatal()
	}

	if v.IsInt() == false {
		t.Fatal()
	}

	i, ok := v.ToInt()

	if i != 2 || !ok {
		t.Fatal()
	}
}

func Test_Script4(t *testing.T) {
	v := script4.Execute()

	if v == nil {
		t.Fatal()
	}

	if v.IsNumber() == false {
		t.Fatal()
	}

	_, ok := v.ToNumber()

	if !ok {
		t.Fatal()
	}
}

func Benchmark_ADD_IN_JS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		script1.Execute()
	}
}

func Benchmark_ADD_BY_JS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		script2.Execute()
	}
}

func Benchmark_ADD_BY_GO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		script3.Execute()
	}
}

func Benchmark_OOXX_IN_JS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		script4.Execute()
	}
}

func Benchmark_OOXX_IN_GO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ooxx(21233, 3452122)
	}
}
