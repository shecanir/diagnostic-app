package main

import (
	"fmt"
)

var colorMap = map[string]string{
	"orange": "\033[38;5;208m", // Orange
	"green": "\033[32m", // Green
	"white": "\033[37m", // White
	"grey": "\033[90m", // Grey
	"red": "\033[31m", // Red
	"blue": "\033[34m", // Blue
	"yellow": "\033[33m", // Yellow
	"reset": "\033[0m", // Reset
}

func printLogo() {
	// ANSI color map
	asciiArt := []struct {
		color string
		text  string
	}{
		{"orange", "\033[39G.,,.\n"},
		{"orange", "\033[36G,i111111;.\n"},
		{"orange", "\033[35G:t17    \\tt,\n"},
		{"orange", "\033[34G.1t0      0:1\n"},
		{"orange", "\033[35Git:      ;ti\n"},
		{"orange", "\033[36Git1;::;1t;\n"},
		{"orange", "\033[37G.:i;;i:.\n"},
		{"green", "\033[14G.,::::,."},
		{"grey", "\033[40Gii"},
		{"orange", "\033[60G.,::::,.\n"},
		{"green", "\033[13G:iiiiiiii:"},
		{"grey", "\033[39G:0G:"},
		{"orange", "\033[58G.it1;;;itt:\n"},
		{"green", "\033[12G:iiiiiiiiii;"},
		{"grey", "\033[39G,0G,"},
		{"orange", "\033[58Git;     .it;\n"},
		{"green", "\033[12G;iiiiiiiiiii."},
		{"grey", "\033[39G.::."},
		{"orange", "\033[57G.t1.      ::i\n"},
		{"green", "\033[12G,iiiiiiiiiii;:,.       .,:;;;;;;:,."},
		{"grey", "\033[56G,"},
		{"orange", "\033[58G;t1,    :tt,\n"},
		{"green", "\033[13G.;iiiiii;;iiiii;:. .:;iii;;:::;iii;:."},
		{"grey", "\033[52G,;tCG;"},
		{"orange", "\033[59G:1t1111ti,\n"},
		{"green", "\033[14G...,..  .,:;iiii;;ii:..       .:iii:"},
		{"grey", "\033[52G1Gfi,"},
		{"orange", "\033[61G.,,,,.\n"},
		{"green", "\033[27G,:;iii;.          . .:ii;"},
		{"grey", "\033[53G.\n"},
		{"green", "\033[29G,ii;          .;i,  ;ii:\n"},
		{"green", "\033[29G;ii,  ,,    ,;i:.   .ii;\n"},
		{"green", "\033[29G;ii,  :i;,,;i:.     .ii;\n"},
		{"green", "\033[29G,ii;   .:ii:.       ;ii:\n"},
		{"grey", "\033[28G."},
		{"green", "\033[30G:ii;.    v       .:ii:."},
		{"grey", "\033[53G.\n"},
		{"orange", "\033[15G.,::,."},
		{"grey", "\033[24G.:1CGi"},
		{"green", "\033[31G:iii:..       .:iii:"},
		{"grey", "\033[52G1GL1:."},
		{"orange", "\033[61G.,::,.\n"},
		{"orange", "\033[13G,1t1ii1t1:"},
		{"grey", "\033[22GiGL1:."},
		{"green", "\033[32G.:;iii;;:::;iii;:."},
		{"grey", "\033[52G.:1CG;"},
		{"orange", "\033[59G:1t1ii1ti,\n"},
		{"orange", "\033[12G:t1,    .1t;"},
		{"grey", "\033[25G."},
		{"green", "\033[35G.,:;;;;;;:,."},
		{"grey", "\033[56G."},
		{"orange", "\033[58G;ti.    ,1t:\n"},
		{"orange", "\033[12G1,,      .t1."},
		{"grey", "\033[39G.;;."},
		{"orange", "\033[57G.t1.      ::i\n"},
		{"orange", "\033[12G;ti.    .it;"},
		{"grey", "\033[39G,0G,"},
		{"orange", "\033[58Git;     .1t:\n"},
		{"orange", "\033[13G:1ti;;itt;"},
		{"grey", "\033[39G:0G,"},
		{"orange", "\033[58G.;t1i;;it1:\n"},
		{"orange", "\033[15G,:;;:,"},
		{"grey", "\033[40Gi;"},
		{"orange", "\033[60G.,;;;:,\n"},
		{"orange", "\033[37G,;iiii;,\n"},
		{"orange", "\033[35G.1ti:,,:iti.\n"},
		{"orange", "\033[34G.1t:      :ti\n"},
		{"orange", "\033[34G.1t,      ,t1\n"},
		{"orange", "\033[35G,tt:.  .;t1,\n"},
		{"orange", "\033[36G.;111111;.\n"},
		{"orange", "\033[39G....\n"},
		{"white", "\n\n \033[24GWelcome to the Shecan Diagnose CLI!\n"},
		{"grey", "\033[36GVersion 0.1\n"},
		{"white", "\n"},
	}

	for _, line := range asciiArt {
		fmt.Printf("%s%s", colorMap[line.color], line.text)
	}
	fmt.Println()
}
