# XmlComparator

GoLang library for comparing XML strings

## Overview

API consists of two functions - 
```
xmlcomparator.CompareXmlStrings(sample1 string, sample2 string, stopOnFirst bool) []string
xmlcomparator.CompareXmlStringsEx(sample1 string, sample2 string, stopOnFirst bool, ignored []string) []string
```
that return a list of detected differences between two XML samples. Comparison can be stopped on the first occasion - `stopOnFirst=true`. The second form takes a list of RegEx strings to be used as a filter to discard differences.

Each entry in the returned list contains XML path to the node like  `..., path='/note/to[0]'`. Path elements might contain zero-based index of an element in the siblings list.

When difference in children nodes is detected, message has form `Children differ: 3 vs 4 (diffs: ...)` where the first number is count of childrent inthe first sample.
Mismatched child elements in `diffs` list have numeric suffix like `:+1` or `:-3`. Positive number relates to count of elements with same name in `sample1` not present in `sample2`, negative - when `sample2` has extra elements.

Example of usage in the code -
```go
    import (
        "github.com/aknopov/xmlcomparator"
    )

    ...
    xmlSample1 := "<a><b/><c/></a>"
	xmlSample2 := "<a><c/><b/></a>"
    diffs := xmlcomparator.CompareXmlStrings(xmlSample1, xmlSample2, false)
    assert.Equal([]string{"Children order differ for 2 nodes, path='/a'"},
        CompareXmlStrings(xmlSample1, xmlSample2, true))
```
See more examples in unit tests.