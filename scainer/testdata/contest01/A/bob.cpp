#include <bits/stdc++.h>
using namespace std;

// Заведомо "списанная" с alice пара: переименованы переменные, добавлен комментарий.
int main() {
    int count;
    cin >> count;
    long long total = 0;
    for (int idx = 0; idx < count; idx++) {
        int value;
        cin >> value;
        total += value;
    }
    cout << total << endl;
    return 0;
}
