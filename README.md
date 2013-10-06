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

(print (+ "Hello " name)) -> Hello gamelisp
````

Data Types
----------

Gamelisp uses a simplified data system with only a hand full of types.

````clojure
(map type [
	"ABC" 
	true 
	42 
	1.5E+6 
	[1 2 3] 
	{:version 1.0}
	:name
	(symbol "x")
	Nothing])

-> [String, Bool, Int, Float, List, Dict, Keyword, Symbol, Nothing]
````

Arithmetic
----------

Basic arithmetic has been implemented for Int and Float. Additionally String supports multiplication like in Python for repeating strings and addition for concatenation. Plus (+) can also be used to join lists or items to lists. As a general rule the + function never modifies its arguments and always returns a new object.

````clojure
(+ 30 5) -> 35
(- 99 5.0) -> 94.0
(* 6 6) -> 36
(/ 8 5) -> 1
(/ 8 5.) -> 1.6
(* "W" 3) -> "WWW" 
(+ "Hello #" (/ 10 2)) -> "Hello #5"
(+ [1 2] 3) -> [1 2 3]
(+ [1 2] [3 4]) -> [1 2 3 4]
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
(slice list startIncl [endExcl]) ; Get a slice of a list, allows negative indices
(append list xs1 [xs2 [...]]) ; Appends lists of items to given list and returns the modified list
(prepend list xs1 [xs2 [...]]); Analogous to append, just prepending instead
(do expr1 expr2 ...) ; evaluate multiple expressions and return the value of the last
````

Function Definition and Multiple Dispatch
-----------------------------------------

Examples:
````clojure
(defn square [x] (* x 2))

; Multiple Dispatch with simple pattern matching
(defn| fib [0] 0)
(defn| fib [1] 1)
(defn| fib [n] (+ (fib (- n 1)) (fib (- n 2)))

(defn| is-string [(x String)] true)
(defn| is-string [x] false)

; Using Anonynmous functions and implicit parameters 
(map #(* 2 %) [1 2 3]) -> [2 4 6]
````

TODOs:
-----------------------------------------

* Runtime Code Reloading
* OpenGL/SDL support