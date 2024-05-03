// Automatically generate a github README.md for each package in your Go module.
//
// # Examples
//
// Create a README.md for each package in the current module (Warning: this will overwrite any existing README.md files!)
//
//	autoreadme
//
// Get a copy of the default template for customization
//
//	autoreadme -print-template >README.md.template
//
// # Template Variables
//
// Templates are executed by Go's text/template. The dot is set to a struct containing (each explained below)
//
//	.ProjectRoot
//	.Repository
//	.Module
//	.Package
//
// ProjectRoot is a boolean which is true if the current package is in the same directory as go.mod.
//
// The Repository entry consists solely of .Data, which may contain optional arbitrary data
// set by JSON in .github/autoreadme/README.md.data.
//
// The Module entry contains:
//
//	.Path - the name of the module
//	.Version - the version of the module
//	.Deprecated - deprecation notice, if present
//	.GoVersion - the language version of the module
//	.Toolchain - the toolchain version of the module
//	.Documentation - a Documentation entry (see below)
//
// The Package entry contains:
//
//	.Name
//	.Import - the import path of the package
//	.Documentation - a Documentation entry (see below)
//	.Data - optional arbitrary data set by a JSON in ./README.md.data
//	.Library - true if not a command
//	.Command - true if package main
//	.Notes - a list of Note (see below) entries
//	.Examples - a map of names to Example (see below) entries
//	.ExternalExamples - like Examples but for examples from package X_test
//
// Name is the package name for libraries. However, for commands that is always "main"
// so for commands it uses the name of the directory.
//
// Documentation entries contain
//
//	.Synopsis - The first sentence as plain text
//	.Doc - the raw go/doc/comment/Doc
//
// additionally Documentation has a
//
//	.ToMarkdown headingLevel
//
// method that renders .Doc as markdown.
// Each heading in Doc is set to headingLevel, allowing them to be properly nested in context.
//
// For packages, the documentation is computed in the standard way by comments attached to the package token.
// For modules, similar logic is used for the comments attached to the module token (note: any Deprecated: comment is not included, but can be accessed from .Module.Deprecated).
//
// Example entries contain
//
//	.Code - markdown formatted code
//	.Output - the expected output, if specified
//	.Playable - if the example is self-contained
//
// Note entries collect the "KIND(uid): body" notes from the package (per go/doc)
//
//	.Kind - the kind of note
//	.UID - the name associated with this name
//	.Body - the text of the note
//
// The list of notes has two methods, Kind and UID, which take a string
// and return a list of notes matching the respective Kind/UID.
//
// # Repository Configuration
//
// A number of files can be added to a .github/autoreadme directory at the root of your repository.
//  1. .github/autoreadme/README.md.template will override the built in default template for all packages without their own template
//  2. .github/autoreadme/README.md.data can contain a single JSON document that can be accessed by every template using .Repository.Data
//  3. .github/autoreadme/autoreadme.ignore can contain a newline separated list of import paths that will be skipped when README.md are generated
//
// # Package configuration
//
// Two files can be added to the directory of each package for one-off customizations
//  1. README.md.template will override the default (or repository) template
//  2. README.md.data can contain a single JSON document that can be accessed by .Package.Data
package main
