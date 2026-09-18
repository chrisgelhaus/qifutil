/*
Copyright © 2025 Chris Gelhaus
*/
package cmd

import "regexp"

// The patterns below are compiled once, at package level. Compiling them in a
// command body meant the error was discarded with _ in several places; a
// pattern that does not compile is a defect in the program rather than
// something a QIF file can cause, and MustCompile fails loudly at the point
// the mistake was made.

// accountBlockPattern matches the header that opens a bank or credit card
// account, capturing its name and type. It was written out identically in
// seven command files.
var accountBlockPattern = regexp.MustCompile(
	`(?m)^!Account[^\n]*\n^N(.*?)\n^T(.*?)\n^\^\n^!Type:(Bank|CCard)\s*\n`)

// anyAccountBlockPattern matches an account header of any type, which
// account-stats uses because it reports on every account rather than only the
// ones holding transactions.
var anyAccountBlockPattern = regexp.MustCompile(
	`!Account\nN(.*?)\nT(.*?)\n\^\n!Type:(.*?)\n`)

// nextTypePattern finds the next !Type line, which bounds the block being read.
var nextTypePattern = regexp.MustCompile(`(?mi)^\s*!Type:.*$`)

// categoryBlockPattern and tagBlockPattern find the start of the category and
// tag lists a QIF file may carry.
var categoryBlockPattern = regexp.MustCompile(`(?m)^!Type:Cat\n`)

var tagBlockPattern = regexp.MustCompile(`(?m)^!Type:Tag\n`)

// categoryRecordPattern and tagRecordPattern match one entry of those lists.
var categoryRecordPattern = regexp.MustCompile(
	`(?m)(^N(.*)\n(^D(.*)\n)?(^T(.*)\n)?(^R(.*)\n)?(^E(.*)\n)?(^I(.*)\n)?^\^\n)`)

var tagRecordPattern = regexp.MustCompile(`(?m)(^N(.*)\n^(D(.*)\n^)?\^\n)`)
