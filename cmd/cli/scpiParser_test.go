package main

import (
	"strings"
	"testing"
)

func BenchmarkVxgScpi(b *testing.B) {
	for i := 0; i < b.N; i++ {
		head := scpiNode{}
		lines, _ := readLinesFromPath("benchmark_SCPI.txt")
		commands := splitScpiCommands(lines)

		for _, command := range commands {
			createScpiTreeBranch(command, &head)
		}
	}
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
	if !strings.HasSuffix(commands[7][len(commands[2])-1].Text, "?") {
		t.Error(":Hello|Goodbye:My:Friend|Love corner case failed, ? not transferred to both options")
	}
	if !strings.HasSuffix(commands[6][len(commands[2])-1].Text, "?") {
		t.Error(":Hello|Goodbye:My:Friend|Love corner case failed, ? not transferred to both options")
	}
}

func TestScpiParserSuffix(t *testing.T) {
	line := ":SOURce:RADio{1:1}:ALL:OFF/nquery/"
	command := reformatSuffixes(line)
	if !strings.Contains(command, "@") {
		t.Error(line, " Not parsed properly")
	}
}
func TestScpiParserIrregularSuffix(t *testing.T) {
	line := ":SOURce1:RADio1:ALL:OFF/nquery/"
	command := reformatIrregularSuffixes(line)
	if !strings.Contains(command, "@") {
		t.Error(line, " Not parsed properly")
	}
}

func TestGenerateTree(t *testing.T) {
	lines, _ := readLinesFromPath("SCPI.txt")
	parseScpi(lines) //TODO: Generate real tests
}
