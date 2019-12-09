local str = "我"
print(#str, str, "我")

for i = 1, #str do

print(string.format("%02X", str:byte(i)))
end