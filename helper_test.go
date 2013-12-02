package gosrv

import (
  "testing"
)


func testAssertEqual(t *testing.T, v1, v2 interface{}) {
  if v1 != v2 { t.Fatalf("Expected %v but was %v", v1, v2) }
}
