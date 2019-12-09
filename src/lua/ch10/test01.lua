local function newCounter()
    local count = 11
    return function()
       count = count + 1
       return count
    end
end
local c1 = newCounter()
print(c1())
print(c1())
print(c1())
--print("hhhhhhhh")
--print(c1())

--local c2 = newCounter()
--print(c2())
--print(c1())
--print(c2())
