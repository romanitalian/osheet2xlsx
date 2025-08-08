package osheet

import "testing"

func TestInferCell(t *testing.T) {
	if c := inferCell("true"); c.Type != ValueBool || !c.BoolValue {
		t.Fatalf("bool true failed")
	}
	if c := inferCell("12"); c.Type != ValueNumber || c.NumberValue != 12 {
		t.Fatalf("number failed")
	}
	if c := inferCell("2024-01-02"); c.Type != ValueDateTime {
		t.Fatalf("date failed")
	}
	if c := inferCell("hello"); c.Type != ValueString || c.StringValue != "hello" {
		t.Fatalf("string failed")
	}
}
