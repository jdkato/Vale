# This is a single comment line

#=
  I am a mlutiline comment, but i use words that are
  obviously maybe avoidable
=#

"""
    f(x)

I am a simple doc string of a quadratic function.
This is Markdown and hence the first line is just a code block
"""
f(x) = x^2

raw"""
    g(x)

I am an example doc string, which is also in Markdown,
just that math is in ``g(x) = x^3`` two backticks
"""
function g(x)
    return x^3
end

# single strings could also be doc strings I think.

function h(x)
    println("I am just a single line string")
    return x^4
end