package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/itsadijmbt/JsonParser/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// / TokenType defines the possible types of tokens in JSON.
type TokenType int

const jsonFile = "data.json"

// ! defined grammer
const (
	TokenObjectStart TokenType = iota ///< {
	TokenObjectEnd                    ///< }
	TokenArrayStart                   ///< [
	TokenArrayEnd                     ///< ]
	TokenColon                        ///< :
	TokenComma                        ///< ,
	TokenString                       ///< string literal
	TokenNumber                       ///< number
	TokenTrue                         ///< true
	TokenFalse                        ///< false
	TokenNull                         ///< null
	TokenEOF                          ///< end of input
)

// / Token represents a single token with its type and optional value.
// / The Value field is a string for TokenString, a float64 for TokenNumber, and nil otherwise.
type Token struct {
	Type  TokenType
	Value interface{}
}

// /**
// * @brief Tokenizes a JSON string into a slice of tokens.
// *
// * @details This function iterates through the input JSON string, identifying and categorizing tokens.
// * It handles:
// * - Whitespace characters (' ', '\t', '\n', '\r') by skipping them as they are insignificant outside strings.
// * - Structural characters ('{', '}', '[', ']', ':', ',') by mapping them to their respective token types.
// * - String literals (starting with '"') by delegating to parseString.
// * - Numbers (starting with digits or '-') by delegating to parseNumber.
// * - Literal 'true' (starting with 't') by checking the full word and creating a TokenTrue token.
// * - Literal 'false' (starting with 'f') by checking the full word and creating a TokenFalse token.
// * - Literal 'null' (starting with 'n') by checking the full word and creating a TokenNull token.
// * - Unexpected characters by returning an error.
// * The function appends a TokenEOF at the end to signify the end of input.
// *
// * @param jsonStr The JSON string to tokenize.
// * @return A slice of tokens and an error (nil if successful).
// */
func tokenize(jsonStr string) ([]Token, error) {
	var tokens []Token
	index := 0
	/// Skip whitespace characters: space (' '), tab ('\t'), newline ('\n'), carriage return ('\r')
	for index < len(jsonStr) {
		char := jsonStr[index]
		if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			index++
			continue
		}
		switch char {
		case '{':
			/// Append token for object start.
			tokens = append(tokens, Token{Type: TokenObjectStart})
			index++
		case '}':
			/// Append token for object end.
			tokens = append(tokens, Token{Type: TokenObjectEnd})
			index++
		case '[':
			/// Append token for array start.
			tokens = append(tokens, Token{Type: TokenArrayStart})
			index++
		case ']':
			/// Append token for array end.
			tokens = append(tokens, Token{Type: TokenArrayEnd})
			index++
		case ':':
			/// Append token for colon.
			tokens = append(tokens, Token{Type: TokenColon})
			index++
		case ',':
			/// Append token for comma.
			tokens = append(tokens, Token{Type: TokenComma})
			index++
		case '"':
			/// Parse string literal.
			str, newIndex, err := parseString(jsonStr, index)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, Token{Type: TokenString, Value: str})
			index = newIndex
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			/// Parse number.
			num, newIndex, err := parseNumber(jsonStr, index)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, Token{Type: TokenNumber, Value: num})
			index = newIndex
		case 't':
			/// Handle literal 'true'.
			if index+4 <= len(jsonStr) && jsonStr[index:index+4] == "true" {
				tokens = append(tokens, Token{Type: TokenTrue})
				index += 4
			} else {
				return nil, fmt.Errorf("invalid token at %d, expected 'true'", index)
			}
		case 'f':
			/// Handle literal 'false'.
			if index+5 <= len(jsonStr) && jsonStr[index:index+5] == "false" {
				tokens = append(tokens, Token{Type: TokenFalse})
				index += 5
			} else {
				return nil, fmt.Errorf("invalid token at %d, expected 'false'", index)
			}
		case 'n':
			/// Handle literal 'null'.
			if index+4 <= len(jsonStr) && jsonStr[index:index+4] == "null" {
				tokens = append(tokens, Token{Type: TokenNull})
				index += 4
			} else {
				return nil, fmt.Errorf("invalid token at %d, expected 'null'", index)
			}
		default:
			return nil, fmt.Errorf("unexpected character at %d: %c", index, char)
		}
	}
	/// Append end-of-file token.
	tokens = append(tokens, Token{Type: TokenEOF})
	return tokens, nil
}

// /**
// * @brief Parses a JSON string literal starting at the given index.
// *
// * @details This function processes a string literal starting with '"', handling:
// * - Normal characters by adding them to the result.
// * - Escape sequences (e.g., '\t', '\n', '\r', '\f', '\b') by interpreting them correctly.
// * - Unicode escapes ('\uXXXX') by converting the hexadecimal code to a rune.
// * - The closing quote ('"') to terminate the string.
// * It returns an error if the string is unterminated or contains invalid escape sequences.
// *
// * @param jsonStr The JSON string being parsed.
// * @param index The current index in the string (should point to the opening quote).
// * @return The parsed string, the new index after the closing quote, and any error.

func parseString(jsonStr string, index int) (string, int, error) {
	if jsonStr[index] != '"' {
		return "", index, fmt.Errorf("expected quote at %d", index)
	}
	index++ // Skip opening quote

	var sb strings.Builder

	for index < len(jsonStr) {
		char := jsonStr[index]

		if char == '"' {
			/// End of string literal.
			return sb.String(), index + 1, nil
		}
		if char == '\\' {
			index++
			if index >= len(jsonStr) {
				return "", index, fmt.Errorf("unterminated string")
			}

			switch jsonStr[index] {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				/// Append the escaped character.
				sb.WriteByte(jsonStr[index])
				index++
			case 'u':
				/// Handle Unicode escape sequence '\uXXXX'.
				if index+4 >= len(jsonStr) {
					return "", index, fmt.Errorf("invalid unicode escape at %d", index)
				}
				hex := jsonStr[index+1 : index+5]
				r, err := strconv.ParseUint(hex, 16, 32)
				if err != nil {
					return "", index, err
				}
				sb.WriteRune(rune(r))
				index += 5
			default:
				return "", index, fmt.Errorf("invalid escape character at %d", index)
			}
		} else {
			/// Append regular character.
			sb.WriteByte(char)
			index++
		}
	}
	return "", index, fmt.Errorf("unterminated string")
}

// /**
// * @brief Parses a JSON number starting at the given index.
// *
// * @details This function parses numbers, which may include:
// * - Integers (e.g., "123").
// * - Floating-point numbers (e.g., "12.34").
// * - Scientific notation (e.g., "1.23e-4").
// * It stops parsing when it encounters a character not part of a number and converts the substring to a float64.
// *
// * @param jsonStr The JSON string being parsed.
// * @param index The current index in the string (should point to the start of the number).
// * @return The parsed number as a float64, the new index, and any error.
// */
func parseNumber(jsonStr string, index int) (float64, int, error) {
	start := index
	for index < len(jsonStr) {
		char := jsonStr[index]
		if (char >= '0' && char <= '9') || char == '.' || char == 'e' || char == 'E' || char == '+' || char == '-' {
			index++
		} else {
			break
		}
	}
	numStr := jsonStr[start:index]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, start, err
	}
	return num, index, nil
}

// / TokenStream manages the sequence of tokens.
type TokenStream struct {
	tokens []Token
	index  int
}

// /**
// * @brief Advances to the next token in the stream.
// *
// * @details Returns the next token or a TokenEOF if the end of the token list is reached.
// *
// * @return The next Token.
// */
func (ts *TokenStream) Next() Token {
	if ts.index < len(ts.tokens) {
		token := ts.tokens[ts.index]
		ts.index++
		return token
	}
	return Token{Type: TokenEOF}
}

// /**
// * @brief Peeks at the next token without advancing the stream.
// *
// * @details Returns the next token or a TokenEOF if the end of the token list is reached.
// *
// * @return The next Token without consuming it.
// */
func (ts *TokenStream) Peek() Token {
	if ts.index < len(ts.tokens) {
		return ts.tokens[ts.index]
	}
	return Token{Type: TokenEOF}
}

// /**
// * @brief Parses a slice of tokens into a Go data structure.
// *
// * @details Initializes a TokenStream and parses the JSON value, ensuring no extra tokens remain after parsing.
// *
// * @param tokens The slice of tokens to parse.
// * @return The parsed JSON value or an error.
// */
func parse(tokens []Token) (interface{}, error) {
	ts := &TokenStream{tokens: tokens, index: 0}
	value, err := parseValue(ts)
	if err != nil {
		return nil, err
	}
	if ts.Peek().Type != TokenEOF {
		return nil, fmt.Errorf("extra tokens after value")
	}
	return value, nil
}

// /**
// * @brief Parses a single JSON value from the token stream.
// *
// * @details Dispatches to specific parsing functions based on the token type:
// * - TokenObjectStart ('{') -> parseObject
// * - TokenArrayStart ('[') -> parseArray
// * - TokenString -> returns the string value
// * - TokenNumber -> returns the float64 value
// * - TokenTrue ('true') -> returns true
// * - TokenFalse ('false') -> returns false
// * - TokenNull ('null') -> returns nil
// *
// * @param ts The TokenStream to read from.
// * @return The parsed value or an error.
// */
func parseValue(ts *TokenStream) (interface{}, error) {
	token := ts.Next()
	switch token.Type {
	case TokenObjectStart:
		return parseObject(ts)
	case TokenArrayStart:
		return parseArray(ts)
	case TokenString:
		return token.Value.(string), nil
	case TokenNumber:
		return token.Value.(float64), nil
	case TokenTrue:
		return true, nil
	case TokenFalse:
		return false, nil
	case TokenNull:
		return nil, nil
	default:
		return nil, fmt.Errorf("unexpected token: %v", token)
	}
}

// /**
// * @brief Parses a JSON object from the token stream.
// *
// * @details Reads key-value pairs until encountering '}', handling:
// * - Commas (',') between pairs (except before the first pair).
// * - Colons (':') between keys and values.
// * - String keys followed by values of any type.
// *
// * @param ts The TokenStream to read from.
// * @return A map representing the object or an error.
// */
func parseObject(ts *TokenStream) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	first := true
	for {
		token := ts.Peek()
		if token.Type == TokenObjectEnd {
			/// Consume the '}' token and return the object.
			ts.Next()
			return obj, nil
		}
		if !first {
			if token.Type != TokenComma {
				return nil, fmt.Errorf("expected ',' or '}'")
			}
			/// Consume the comma.
			ts.Next()
			token = ts.Peek()
		}
		if token.Type != TokenString {
			return nil, fmt.Errorf("expected string key")
		}
		/// Get the key.
		key := ts.Next().Value.(string)
		if ts.Next().Type != TokenColon {
			return nil, fmt.Errorf("expected ':'")
		}
		/// Parse the value.
		value, err := parseValue(ts)
		if err != nil {
			return nil, err
		}
		obj[key] = value
		first = false
	}
}

// /**
// * @brief Parses a JSON array from the token stream.
// *
// * @details Reads values until encountering ']', handling:
// * - Commas (',') between values (except before the first value).
// * - Values of any type (objects, arrays, strings, numbers, true, false, null).
// *
// * @param ts The TokenStream to read from.
// * @return A slice representing the array or an error.
// */
func parseArray(ts *TokenStream) ([]interface{}, error) {
	var arr []interface{}
	first := true
	for {
		token := ts.Peek()
		if token.Type == TokenArrayEnd {
			/// Consume the ']' token and return the array.
			ts.Next()
			return arr, nil
		}
		if !first {
			if token.Type != TokenComma {
				return nil, fmt.Errorf("expected ',' or ']'")
			}
			/// Consume the comma.
			ts.Next()
		}
		/// Parse the value.
		value, err := parseValue(ts)
		if err != nil {
			return nil, err
		}
		arr = append(arr, value)
		first = false
	}
}

// /**
// * @brief Main entry point to parse a JSON string into a Go data structure.
// *
// * @details Combines tokenization and parsing:
// * - Calls tokenize to break the JSON string into tokens.
// * - Calls parse to convert tokens into a Go value.
// *
// * @param jsonStr The JSON string to parse.
// * @return The parsed JSON value or an error.
// */
func ParseJSON(jsonStr string) (interface{}, error) {
	tokens, err := tokenize(jsonStr)
	if err != nil {
		return nil, err
	}
	return parse(tokens)
}

// /**
// * @brief Escapes special characters in a string for JSON output.
// *
// * @param s The string to escape.
// * @return The escaped string enclosed in quotes.
// */
func escapeString(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			/// Escape double quotes.
			sb.WriteString("\\\"")
		case '\\':
			/// Escape backslashes.
			sb.WriteString("\\\\")
		case '\b':
			/// Escape backspace.
			sb.WriteString("\\b")
		case '\f':
			/// Escape formfeed.
			sb.WriteString("\\f")
		case '\n':
			/// Escape newline.
			sb.WriteString("\\n")
		case '\r':
			/// Escape carriage return.
			sb.WriteString("\\r")
		case '\t':
			/// Escape tab.
			sb.WriteString("\\t")
		default:
			if r < 32 || r > 126 {
				/// Escape non-printable characters.
				sb.WriteString(fmt.Sprintf("\\u%04x", r))
			} else {
				sb.WriteRune(r)
			}
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

// /**
// * @brief Recursively formats the JSON value with indentation.
// *
// * @param value The JSON value to format.
// * @param sb The string builder to append the formatted text.
// * @param indentLevel The current level of indentation.
// */
func prettyPrint(value interface{}, sb *strings.Builder, indentLevel int) {
	indent := strings.Repeat(" ", indentLevel*2)
	switch v := value.(type) {
	case map[string]interface{}:
		sb.WriteString("{\n")
		first := true
		for key, val := range v {
			if !first {
				sb.WriteString(",\n")
			}
			sb.WriteString(indent + "  \"" + key + "\": ")
			prettyPrint(val, sb, indentLevel+1)
			first = false
		}
		sb.WriteString("\n" + indent + "}")
	case []interface{}:
		sb.WriteString("[\n")
		first := true
		for _, val := range v {
			if !first {
				sb.WriteString(",\n")
			}
			sb.WriteString(indent + "  ")
			prettyPrint(val, sb, indentLevel+1)
			first = false
		}
		sb.WriteString("\n" + indent + "]")
	case string:
		sb.WriteString(escapeString(v))
	case float64:
		sb.WriteString(fmt.Sprintf("%v", v))
	case bool:
		if v {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case nil:
		sb.WriteString("null")
	default:
		sb.WriteString("unknown type")
	}
}

// /**
// * @brief Formats the JSON value into a pretty-printed string.
// *
// * @param jsonValue The JSON value to format.
// * @return A pretty-printed JSON string.
// */
func PrettyPrint(jsonValue interface{}) string {
	var sb strings.Builder
	prettyPrint(jsonValue, &sb, 0)
	return sb.String()
}

// /**
// * @brief Recursively processes nested JSON strings.
// *
// * @details This function traverses the JSON data structure and, for every string that appears to be valid JSON (i.e.,
// * starting with '{' or '['), it attempts to parse it as JSON and replaces the string with the parsed value.
// * This is done recursively to account for multiple levels of nested JSON.
// *
// * @param value The JSON value to process.
// * @return The processed JSON value with nested JSON parsed.
// */
func processNestedJSON(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		for k, val := range v {
			v[k] = processNestedJSON(val)
		}
		return v
	case []interface{}:
		for i, val := range v {
			v[i] = processNestedJSON(val)
		}
		return v
	case string:
		trimmed := strings.TrimSpace(v)
		if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
			// Attempt to parse the string as JSON.
			nested, err := ParseJSON(v)
			if err == nil {
				return processNestedJSON(nested)
			}
		}
		return v
	default:
		return v
	}
}

func main() {
	/// Example JSON string that includes nested JSON as a string.
	f, err := os.OpenFile(jsonFile, os.O_RDWR, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open %s: %v\n", jsonFile, err)
		os.Exit(1)
	}
	defer f.Close()

	info, _ := f.Stat()
	if info.Size() == 0 {
		f.WriteString("// Paste your JSON here and save\n")
		fmt.Printf("Please add your JSON to %s and run again.\n", jsonFile)
		return
	}
	data, err := io.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
		os.Exit(1)
	}
	result, err := ParseJSON(string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}
	tree := processNestedJSON(result)

	if err := tea.NewProgram(ui.NewModel(tree)).Start(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}

}
