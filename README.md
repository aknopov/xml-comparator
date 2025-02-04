![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/aknopov/xml-comparator/go.yml)
![Coveralls](https://img.shields.io/coverallsCoverage/github/aknopov/xml-comparator)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/google.golang.org/xmlcomparator.svg)](https://pkg.go.dev/github.com/aknopov/xmlcomparator)

# XmlComparator

GoLang library for comparing XML strings

## Overview

API consists of two functions - 
```
xmlcomparator.CompareXmlStrings(sample1 string, sample2 string, stopOnFirst bool) []string
xmlcomparator.CompareXmlStringsEx(sample1 string, sample2 string, stopOnFirst bool, ignored []string) []string
```
that return a list of detected differences between two XML samples. Comparison can be stopped on the first occasion - `stopOnFirst=true`. The second form takes a list of RegEx strings to be used as a filter for ignored differences.

Each entry in the returned list contains the XML path to the node like  `..., path='/note/to[0]'`. Path elements might contain zero-based index of an element in the siblings list.

When a difference in children elements is detected, the message has the form `Children differ: counts 3 vs 4: ...` where the first number is the count of children in the first sample.
Mismatched child elements in the `diffs` list have two numbers. The first, in square brackets, is the index in the sibling nodes list.
The second - suffix like `:+1` or `:-3` is the count of consecutive mismatched elements with the same name. A positive number relates to the count of elements in `sample1`, negative - to `sample2`.

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

    xmlSample3 := `<a><b><c/><c/><d/></b></a>`
    xmlSample4 := `<a><b><d/><e/><e/><e/></b></a>`
    assert.Equal([]string{"Children differ: counts 3 vs 4: c[0]:+2, e[1]:-3, path='/a/b'"}, CompareXmlStrings(xmlSample3, xmlSample4, false))

    diffs := CompareXmlStringsEx(xmlString1, xmlMixed, false, []string{`Node texts differ: '.+' vs '.+'`})
    assert.Equal(1, len(diffs))

    xmlString5 := `<a>Node Content</a>`
    xmlString6 := `<a>Another Content</a>`
    diffs = CompareXmlStringsEx(xmlString5, xmlString6, false, []string{`Node textsNodes test differ: '.+' vs '.+'`})
    assert.Equal(0, len(diffs))
```