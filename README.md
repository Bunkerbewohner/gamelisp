gamelisp
========

My game-oriented programming language inspired by Lisp, Go, Python, Clojure; implemented in Go. I intend to use it as a DSL for creating small 3D games using SDL2.

The language allows both imperative and functional programming. Lists and dictionaries are mutable by default.

This project started September 23rd 2013, is under active development and lacking many features. 

Syntax
------

The syntax is inspired by LISP and Clojure.

````clojure
(def x 5) -> 5 ; defines a *new* variable with value 5
(def! x 2) -> 2 ; overwrites existing variable value
(def name "gamelisp") -> "gamelisp"

(print (+ "Hello " name) -> Hello gamelisp
````

Data Types
----------

Gamelisp uses a simplyfied data system with only a hand full of types.

````clojure
(map type [
	"ABC" 
	true 
	42 
	1.5E+6 
	[1 2 3] 
	{:version 1.0}
	:name
	(symbol x)
	Nothing])

-> ["String", "Bool", "Int", "Float", "List", "Dict", "Keyword", "Symbol", "Nothing"]
````

Arithmetic
----------

Basic arithmetic has been implemented for Int and Float. Additionally String supports multiplication like in Python for repeating strings and addition for concatenation.

````clojure
(+ 30 5) -> 35
(- 99 5.0) -> 94.0
(* 6 6) -> 36
(/ 8 5) -> 1
(/ 8 5.) -> 1.6
(* "W" 3) -> "WWW" 
(+ "Hello #" (/ 10 2)) -> "Hello #5"
````

Core Functions (so far)
--------------

````clojure
(def symbol value) ; Define a new variable
(type x) ; Get type of variable
(str x) ; Get string representation of variable
(print x) ; Print value of x
(map f xs) ; Apply function to items in list xs
(filter f xs) ; Filter items xs with function f
(get dict key) ; Get dictionary entry
(get list index) ; Get list entry (negative indices like in Python)
(put dict key value) ; Add or set dictionary entry
(put list index value) ; Set list entry
(len dictOrList) ; Get the length of dict, list or string
(slice list startIncl [endExcl]) ; Get a slice of a list
````