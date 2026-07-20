import sys

data = sys.stdin.read().split()
count = int(data[0])
answer = min(int(v) for v in data[1:count + 1])
print(answer)
