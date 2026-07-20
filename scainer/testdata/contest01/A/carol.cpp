#include <iostream>
#include <vector>
#include <numeric>
using namespace std;

// Независимое решение: другой стиль, STL-алгоритм.
int main() {
    int n;
    cin >> n;
    vector<int> a(n);
    for (auto &v : a) cin >> v;
    cout << accumulate(a.begin(), a.end(), 0LL) << "\n";
}
