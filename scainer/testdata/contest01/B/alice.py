n = int(input())
best = -10**18
for _ in range(n):
    x = int(input())
    best = max(best, x)
print(best)
