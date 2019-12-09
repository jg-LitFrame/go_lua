function div0(a, b)
    error("DIV BY ZERO !")
end
local ok, err = pcall(div0)