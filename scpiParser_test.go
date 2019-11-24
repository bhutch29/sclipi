package main

import (
	"strings"
	"testing"
)

func BenchmarkMxgScpi(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lines, _ := readLines("SCPI.txt")
		parseScpi(lines)
	}
}

func TestGenerateTree(t *testing.T) {
	lines, _ := readLines("SCPI.txt")
	parseScpi(lines) //TODO: Generate real tests
}

func TestScpiParserTwoOptionals(t *testing.T) {
	lines := []string{":DIAGnostic[:CPU]:BLOCk:ABUS:LIST[:SINGle]"}
	commands := splitScpiCommands(lines)
	if len(commands) != 8 {
		t.Error(":DIAGnostic[:CPU]:BLOCk:ABUS:LIST[:SINGle] not parsed properly:", commands)
		return
	}
}

func TestScpiParserThreeOptionals(t *testing.T) {
	lines := []string{"[:SOURce]:AMPLitude[:LEVel]:STEP[:INCRement]"}
	commands := splitScpiCommands(lines)
	if len(commands) != 16 {
		t.Error("[:SOURce]:AMPLitude[:LEVel]:STEP[:INCRement] not parsed properly:", commands)
		return
	}
}

func TestScpiParserFourOptionals(t *testing.T) {
	lines := []string{"[:SOURce]:FREQuency[:CW][:FIXed][:FIXed]"}
	commands := splitScpiCommands(lines)
	if len(commands) != 32 {
		t.Error("[:SOURce]:FREQuency[:CW][:FIXed][:FIXed] not parsed properly:", commands)
		return
	}
}

func TestScpiParserSuffix(t *testing.T) {
	lines := []string{":Example{1:2}:Afterward"}
	commands := splitScpiCommands(lines)
	if len(commands) != 4 {
		t.Error(":Example{1:2}:Afterward not parsed properly:", commands)
		return
	}
	if len(commands[0]) != 2 {
		t.Error(":Example{1:2}:Afterward not parsed properly:", commands[0])
	}
}

func TestScpiParserFirstOptional(t *testing.T) {
	lines := []string{"[:SOURce]:FREQuency:SPAN"}
	commands := splitScpiCommands(lines)
	if len(commands) != 4 {
		t.Error("[:SOURce]:FREQuency:SPAN not parsed properly:", commands)
		return
	}
}

func TestScpiParserNoQuery(t *testing.T) {

	lines := []string{":ABORt/nquery/"}
	commands := splitScpiCommands(lines)
	if len(commands) != 1 {
		t.Error(":ABORt/nquery/ not parsed properly:", commands[0])
		return
	}
	if len(commands[0]) != 1 {
		t.Error(":ABORt/nquery/ not parsed properly:", commands[0])
	}
}

func TestScpiParserOptionalsNoQuery(t *testing.T) {
	lines := []string{":ABORt[:SWEep]/nquery/"}
	commands := splitScpiCommands(lines)
	if len(commands) != 2 {
		t.Error(":ABORt[:SWEep]/nquery/ not parsed properly:", commands)
		return
	}
	if len(commands[0]) != 2 {
		t.Error(":ABORt[:SWEep]/nquery/ not parsed properly:", commands[0])
	}
	if len(commands[1]) != 1 {
		t.Error(":ABORt[:SWEep]/nquery/ not parsed properly:", commands[1])
	}
}

func TestScpiParserOptionals(t *testing.T) {
	lines := []string{":ABORt[:SWEep]"}
	commands := splitScpiCommands(lines)
	if len(commands) != 4 {
		t.Error(":ABORt[:SWEep] not parsed properly:", commands)
		return
	}
	if len(commands[0]) != 2 {
		t.Error(":ABORt[:SWEep] not parsed properly:", commands[0])
	}
	if len(commands[2]) != 1 {
		t.Error(":ABORt[:SWEep] not parsed properly:", commands[2])
	}
}

func TestScpiParserBasic(t *testing.T) {
	lines := []string{":CALibration:BBG:CHANnel:OFFSet"}
	commands := splitScpiCommands(lines)
	if len(commands) != 2 {
		t.Error(":CALibration:BBG:CHANnel:OFFSet not parsed properly:", commands)
		return
	}
	if len(commands[0]) != 4 {
		t.Error(":CALibration:BBG:CHANnel:OFFSet not parsed properly:", commands[0])
	}
	if len(commands[1]) != 4 {
		t.Error(":CALibration:BBG:CHANnel:OFFSet not parsed properly:", commands[1])
	}
}

func TestScpiParserOneBar(t *testing.T) {
	lines := []string{"Hello|Goodbye:My:Friend/nquery/"}
	commands := splitScpiCommands(lines)
	if len(commands) != 2 {
		t.Error(":Hello|Goodbye:My:Friend/nquery not parsed properly")
	}
	if len(commands[0]) != 3 {
		t.Error(":Hello|Goodbye:My:Friend/nquery not parsed properly")
	}
}

func TestScpiParserMultipleBarsNoQuery(t *testing.T) {
	lines := []string{"Hello|Goodbye:My:Friend|Love/nquery/"}
	commands := splitScpiCommands(lines)
	if len(commands) != 4 {
		t.Error(":Hello|Goodbye:My:Friend|Love/nquery not parsed properly")
	}
	if len(commands[0]) != 3 {
		t.Error(":Hello|Goodbye:My:Friend|Love/nquery not parsed properly")
	}
}

func TestScpiParserMultipleBarsCommandAndQuery(t *testing.T) {
	lines := []string{"Hello|Goodbye:My:Friend|Love"}
	commands := splitScpiCommands(lines)
	if len(commands) != 8 {
		t.Error(":Hello|Goodbye:My:Friend|Love not parsed properly")
	}
	if len(commands[0]) != 3 {
		t.Error(":Hello|Goodbye:My:Friend|Love not parsed properly")
	}
	if !strings.HasSuffix(commands[7][len(commands[2]) - 1], "?") {
		t.Error(":Hello|Goodbye:My:Friend|Love corner case failed, ? not transferred to both options")
	}
	if !strings.HasSuffix(commands[6][len(commands[2]) - 1], "?") {
		t.Error(":Hello|Goodbye:My:Friend|Love corner case failed, ? not transferred to both options")
	}
}

func TestBranchSuffixes(t *testing.T) {
	result := handleSuffixes("Example{1:1}")
	if len(result) != 1 {
		t.Error("Example{1:1} not parsed to 1 result", result)
	}

	result = handleSuffixes("Example{1:3}")
	if len(result) != 3 {
		t.Error("Example{1:3} not parsed to 3 results", result)
	}
	if result[0] != "Example1" {
		t.Error("Example{1:3} first element not Example1", result)
	}

	result = handleSuffixes(":Hello{1:2}:World{1:3}:Again{1:2}")
	if len(result) != 12 {
		t.Error(":Hello{1:2}:World{1:3}:Again{1:2} not parsed to 12 results", result)
	}

	result = handleSuffixes("	[:SOURce]:FSIMulator{1:1}:CORRelation:FADer{1:1}:FADer{1:1}:PATH{1:24}")
	if len(result) != 24 {
		t.Error("	[:SOURce]:FSIMulator{1:1}:CORRelation:FADer{1:1}:FADer{1:1}:PATH{1:24} not parsed to 24 results", len(result), result)
	}
}

