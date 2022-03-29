Hello, I am Stan Marsh. This project is my toy interpreted language, common usage supported.
### usage:
commandline mode:
```
stang
```
parse multiple sourcecode files
```
stang [filename...]
```
Note that in the latter mode, the program will exit if the result cannot be evaluated in 3 seconds.
### examples:
#### 1.data types
```
let a = 1
let b = 0.5
let c = true
let d = "hello"
let f = [1, 2, 3]
let g = {name:'stan'}
let h = function(x,y) { return x+y }
print(typeof a, typeof b, typeof c, typeof d, typeof e, typeof f, typeof g, typeof h)
```
#### output:
```
INTEGER, FLOAT, BOOLEAN, STRING, ERROR, ARRAY, HASH, FUNCTION
```

#### 2.operation:
```
print(1+2.5)
print("stan"+"marsh"+1)
print(1||false)
print(!0&&true)

let a = 1
a = 2
print(--a)
delete a
print(a)

let h = {name:'stan', arr:[1,2,3]}
h['age'] = 8
delete h['arr'][0]
delete h['name']
print(h)
```
#### output:
```
3.5
stanmarsh1
true
true
1
2
0
0
Error: unknown identifier: 'a' is not defined
{arr:[1, 2, 3], name:stan}
8
1
stan
{age:8, arr:[null, 2, 3]}
```
The supported operators are:  
`+` `-` `*` `/` `%`  
`&&` `||` `!`  
`+=` `-=` `*=` `/=`  
`++(prefix, suffix)` `--(prefix, suffix)`  
`==` `!=` `>` `<` `>=` `<=`  
`=(assign)` `=(define)`  
`() for function call`  
`[] for array or hash indexing`  
`. for method call`  
`typeof` `delete`

note that in Stang, all expression has a return value.  
The return value of the program is the return value of the last statement or the last 'return' statement, if that exists.
#### 3.function
```
let f = function(x,f){
    return f(f(x))
}
f(2, function(x){return x*x})
```
#### output:
```
16
```
function is a first-class object in Stang.
#### 4.control flow
```
let f = function(x) {
    if (x>0) {
        return true
    } else {
        return false
    }
}
for (let i=-5;i<=5;i+=2) {
    print(f(i))
}

let i = -5
while (true) {
    print(f(i))
    i += 2
    if (i>5) {
        break
    }
}
```
#### output:
```
false
false           
false           
true            
true            
true            
false           
false           
false           
true            
true            
true 
```
#### 5.methods and builtin functions
```
print("stan,kyle".split(','))
print("stan".toUpper().toLower())

print(typeof string(0))
print(number("1.5")+2)

print(now())

let arr = [1,2,3]
arr.push(4)
print(arr.pop())
print(arr)
```
#### output:
```
[stan, kyle]
stan               
STRING             
3.5                
2022-03-30 04:35:06
4                  
[1, 2, 3] 
```