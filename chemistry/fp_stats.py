# %%
from pathlib import Path

import seaborn
seaborn.set_theme()


def get_fp_stats(fps_path: Path):
    fp_dist = 167*[0]
    above_lim = 167*[0]
    total_ones = 0
    max_ones = 0
    n = 0
    with fps_path.open("r") as fd:
        for fp in fd.readlines():
            num_ones = len([x for x in fp if x == '1'])

            total_ones += num_ones
            max_ones = max(num_ones, max_ones)
            fp_dist[num_ones] += 1

            if num_ones == max_ones:
                print(f'Compound[{n}]: {num_ones} -> Total: {total_ones}, Max:{max_ones}')
            n += 1
            # if n > 100000:
            #     break

    print(f'{n} compounds analyzed.  Total: {total_ones}, Max:{max_ones}')
    for i in range (160, 0, -1):
        above_lim[i] = above_lim[i+1] + fp_dist[i]
        print(f"{i}: #comp:{fp_dist[i]} - above: {above_lim[i]} %{above_lim[i]/n}")

    seaborn.relplot(x = range(167), y = fp_dist)

# %%
get_fp_stats(Path.cwd() / "fps.txt")
