/*
 * Copyright 2018 Florent Biville (@fbiville)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"github.com/fbiville/headache/versioning"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"regexp"
	"testing"
)

func TestHeaderWrite(t *testing.T) {
	I := NewGomegaWithT(t)
	newHeader := `// some multi-line header
// with some text`
	regex, _ := computeDetectionRegex([]string{"some multi-line header", "with some text"}, map[string]string{})
	configuration := configuration{
		HeaderRegex:    regexp.MustCompile(regex),
		HeaderContents: newHeader,
		vcsChanges:     []versioning.FileChange{{Path: "../fixtures/hello_world.txt", ReferenceContent: ""}},
	}

	file, err := DryRun(&configuration)

	I.Expect(err).To(BeNil())
	I.Expect(readFile(file)).To(Equal(`file:../fixtures/hello_world.txt
---
0a1,3
> // some multi-line header
> // with some text
> 
---
`))
}
func TestHeaderDoesNotWriteTwice(t *testing.T) {
	I := NewGomegaWithT(t)
	regex, _ := computeDetectionRegex([]string{"some multi-line header", "with some text"}, map[string]string{})
	configuration := configuration{
		HeaderRegex: regexp.MustCompile(regex),
		HeaderContents: `// some multi-line header
// with some text`,
		vcsChanges: []versioning.FileChange{{Path: "../fixtures/hello_world_with_header.txt", ReferenceContent: ""}},
	}

	file, err := DryRun(&configuration)

	I.Expect(err).To(BeNil())
	I.Expect(readFile(file)).To(BeEmpty(), "it should rewrite the file as is")
}

func TestHeaderCommentUpdate(t *testing.T) {
	I := NewGomegaWithT(t)
	regex, _ := computeDetectionRegex([]string{"some multi-line header", "with some text"}, map[string]string{})
	configuration := configuration{
		HeaderRegex: regexp.MustCompile(regex),
		HeaderContents: `/*
 * some multi-line header
 * with some text
 */`,
		vcsChanges: []versioning.FileChange{{Path: "../fixtures/hello_world_with_header.txt", ReferenceContent: ""}},
	}

	file, err := DryRun(&configuration)

	I.Expect(err).To(BeNil())
	I.Expect(readFile(file)).To(Equal(`file:../fixtures/hello_world_with_header.txt
---
1,2c1,4
< // some multi-line header
< // with some text
---
> /*
>  * some multi-line header
>  * with some text
>  */
---
`), "it should rewrite the file with slashstar style")
}

func TestHeaderDataUpdate(t *testing.T) {
	I := NewGomegaWithT(t)
	regex, _ := computeDetectionRegex([]string{"some multi-line header 2017", "with some text from {{.Company}}"},
		map[string]string{
			"Company": "Pairing Corp",
		})
	configuration := configuration{
		HeaderRegex: regexp.MustCompile(regex),
		HeaderContents: `// some multi-line header 2017
// with some text from Pairing Corp`,
		vcsChanges: []versioning.FileChange{{Path: "../fixtures/hello_world_with_parameterized_header.txt", ReferenceContent: ""}},
	}

	file, err := DryRun(&configuration)

	I.Expect(err).To(BeNil())
	I.Expect(readFile(file)).To(Equal(`file:../fixtures/hello_world_with_parameterized_header.txt
---
2c2,3
< // with some text from Soloing Inc.
---
> // with some text from Pairing Corp
> 
---
`))
}

func TestInsertCreationYearAutomatically(t *testing.T) {
	I := NewGomegaWithT(t)
	regex, _ := computeDetectionRegex([]string{"some multi-line header {{.Year}}", "with some text from {{.Company}}"},
		map[string]string{
			"Year":    "{{.Year}}",
			"Company": "Pairing Corp",
		})
	configuration := configuration{
		HeaderRegex: regexp.MustCompile(regex),
		HeaderContents: `// some multi-line header {{.Year}}
// with some text from Pairing Corp`,
		vcsChanges: []versioning.FileChange{{
			Path:             "../fixtures/hello_world.txt",
			CreationYear:     2022,
			ReferenceContent: "",
		}},
	}

	file, err := DryRun(&configuration)

	I.Expect(err).To(BeNil())
	I.Expect(readFile(file)).To(Equal(`file:../fixtures/hello_world.txt
---
0a1,3
> // some multi-line header 2022
> // with some text from Pairing Corp
> 
---
`))
}

func TestInsertCreationAndLastEditionYearsAutomatically(t *testing.T) {
	I := NewGomegaWithT(t)
	regex, _ := computeDetectionRegex([]string{"some multi-line header {{.Year}}", "with some text from {{.Company}}"},
		map[string]string{
			"Year":    "{{.Year}}",
			"Company": "Pairing Corp",
		})
	configuration := configuration{
		HeaderRegex: regexp.MustCompile(regex),
		HeaderContents: `// some multi-line header {{.Year}}
// with some text from Pairing Corp`,
		vcsChanges: []versioning.FileChange{{
			Path:             "../fixtures/hello_world.txt",
			CreationYear:     2022,
			LastEditionYear:  2034,
			ReferenceContent: "",
		}},
	}

	file, err := DryRun(&configuration)

	I.Expect(err).To(BeNil())
	I.Expect(readFile(file)).To(Equal(`file:../fixtures/hello_world.txt
---
0a1,3
> // some multi-line header 2022-2034
> // with some text from Pairing Corp
> 
---
`))
}

func TestDoesNotInsertLastEditionYearWhenEqualToCreationYear(t *testing.T) {
	I := NewGomegaWithT(t)
	regex, _ := computeDetectionRegex([]string{"some multi-line header {{.Year}}", "with some text from {{.Company}}"},
		map[string]string{
			"Year":    "{{.Year}}",
			"Company": "Pairing Corp",
		})
	configuration := configuration{
		HeaderRegex: regexp.MustCompile(regex),
		HeaderContents: `// some multi-line header {{.Year}}
// with some text from Pairing Corp`,
		vcsChanges: []versioning.FileChange{{
			Path:             "../fixtures/hello_world.txt",
			CreationYear:     2022,
			LastEditionYear:  2022,
			ReferenceContent: "",
		}},
	}

	file, err := DryRun(&configuration)

	I.Expect(err).To(BeNil())
	I.Expect(readFile(file)).To(Equal(`file:../fixtures/hello_world.txt
---
0a1,3
> // some multi-line header 2022
> // with some text from Pairing Corp
> 
---
`))
}

func TestHeaderDryRunOnSeveralFiles(t *testing.T) {
	I := NewGomegaWithT(t)
	regex, _ := computeDetectionRegex([]string{"some header {{.Year}}"},
		map[string]string{
			"Year": "{{.Year}}",
		})
	file, err := DryRun(&configuration{
		HeaderRegex:    regexp.MustCompile(regex),
		HeaderContents: "// some header {{.Year}}",
		vcsChanges: []versioning.FileChange{{
			Path:             "../fixtures/hello_world.txt",
			CreationYear:     2022,
			LastEditionYear:  2022,
			ReferenceContent: "",
		}, {
			Path:             "../fixtures/bonjour_world.txt",
			CreationYear:     2019,
			LastEditionYear:  2021,
			ReferenceContent: "",
		}},
	})

	I.Expect(err).To(BeNil())
	I.Expect(readFile(file)).To(Equal(`file:../fixtures/hello_world.txt
---
0a1,2
> // some header 2022
> 
---
file:../fixtures/bonjour_world.txt
---
0a1,2
> // some header 2019-2021
> 
---
`))

}

func TestSimilarHeaderReplacement(t *testing.T) {
	I := NewGomegaWithT(t)
	regex, _ := computeDetectionRegex([]string{"some header {{.Year}} and stuff"},
		map[string]string{
			"Year": "{{.Year}}",
		})
	file, err := DryRun(&configuration{
		HeaderRegex:    regexp.MustCompile(regex),
		HeaderContents: "// some header {{.Year}} and stuff",
		vcsChanges: []versioning.FileChange{{
			Path:             "../fixtures/hello_world_similar.txt",
			CreationYear:     2022,
			LastEditionYear:  2022,
			ReferenceContent: "",
		}},
	})

	I.Expect(err).To(BeNil())
	s := readFile(file)
	I.Expect(s).To(Equal(`file:../fixtures/hello_world_similar.txt
---
1,4c1
< /*
<  *   Some Header 2022 and stuff .
<  *
<  */
---
> // some header 2022 and stuff
---
`))
}

func TestPreserveYear(t *testing.T) {
	I := NewGomegaWithT(t)
	regex, _ := computeDetectionRegex([]string{"Copyright {{.Year}} {{.Company}}"},
		map[string]string{
			"Year": "{{.Year}}",
			"Company": "ACME",
		})
	file, err := DryRun(&configuration{
		HeaderRegex:    regexp.MustCompile(regex),
		HeaderContents: "// some header {{.Year}} {{.Company}}",
		vcsChanges: []versioning.FileChange{{
			Path:             "../fixtures/hello_world_2014.txt",
			CreationYear:     2016,
			LastEditionYear:  2022,
			ReferenceContent: "",
		}},
	})

	I.Expect(err).To(BeNil())
	I.Expect(readFile(file)).To(Equal(`file:../fixtures/hello_world_2014.txt
---
1c1
< // Copyright 2014 ACME
---
> // some header 2014-2022 
---
`))
}

func readFile(file string) string {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}