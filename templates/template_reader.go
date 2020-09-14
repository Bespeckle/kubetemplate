package templates

import (
    "bytes"
    "crypto/rand"
    "math/big"
    "path/filepath"
    "text/template"
)

// Read reads in the contents of filePath, and processes it as a template, returning the processed results as a raw byte
// slice
func Read(filePath string, availableData interface{}) ([]byte, error) {
    tmpl, err := template.New(filepath.Base(filePath)).
        Funcs(availableFunctions).
        ParseFiles(filePath)
    if err != nil {
        return nil, err
    }
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, availableData); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

// availableFunctions is the registry of functions available to use in your YAML templates.
var availableFunctions = template.FuncMap{
    "GeneratePassword": GeneratePassword,
}

var (
    lowerCharSet   = []rune("abcdedfghijklmnopqrst")
    upperCharSet   = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
    specialCharSet = []rune("!@#$%&*")
    numberSet      = []rune("0123456789")
    allCharSet     = append(append(append(append([]rune{}, lowerCharSet...), upperCharSet...), specialCharSet...), numberSet...)
    maxV           = big.NewInt(int64(len(allCharSet)))
)

// passwordLength is the length of passwords GeneratePassword will output.
const passwordLength = 16

// GeneratePassword creates a password and prints it out to standard out on launch.
func GeneratePassword() (string, error) {
    ret := make([]rune, 0, passwordLength)
    // First char needs to be an upper or lower case letter, otherwise it will violate YAML syntax.
    v, err := rand.Int(rand.Reader, maxV)
    if err != nil {
        return "", err
    }
    at := int(v.Uint64())
    at = at % (len(lowerCharSet) + len(upperCharSet))
    ret = append(ret, allCharSet[at])

    // The rest of the characters can be anything.
    for i := 0; i < passwordLength-1; i++ {
        v, err := rand.Int(rand.Reader, maxV)
        if err != nil {
            return "", err
        }
        ret = append(ret, allCharSet[int(v.Uint64())])
    }
    return string(ret), nil
}