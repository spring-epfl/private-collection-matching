# %%
import json
import pprint


import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt
import matplotlib

sns.set_style()
plt.rc('text', usetex=True)
plt.rc('font', family='serif', size=16)
plt.rc('figure', figsize=(5.5,3.5))
plt.rc('text.latex', preamble=r'\usepackage{mathptmx}')



pp = pprint.PrettyPrinter(indent=4)

# %%
def read_bench(file_addr):
    with open(file_addr, 'r') as fd:
        raw = json.load(fd)

    data = {}
    for field in raw[0][0]:
        data[field] = []
    for exps in raw:
        for exp in exps:
            for field in exp:
                data[field].append(exp[field])

    df = pd.DataFrame.from_records(data)
    return df




# %%
# Plot document search
def plot_doc_search(data):
    fig, ax = plt.subplots()

    ax.set_yscale('log')
    ax.set_ylabel('Time (s)')

    plt.xscale('log')
    ax.set_xlabel('\#Documents')
    ticks = [1, 2, 8, 32, 128, 512, 2048, 8192]
    plt.xticks(ticks, ticks)

    df = data.groupby(['SetNum', 'system']).agg(['mean', 'sem'])
    df = df.reset_index()


    comp_format = {
        'CA1' : {
            'label':"CA-MS",
            'color':'#1f78b4',
            'fmt':'o-'
        },
        'CA2' : {
            'label':"CA2",
            'color':'#a6cee3',
            'fmt':'v-'
        },
        'X1' : {
            'label':"X-MS",
            'color':'#33a02c',
            'fmt':'*-'
        },
        'X2' : {
            'label':"X2",
            'color':'#b2df8a',
            'fmt':'^-'
        }
    }

    for system in ["X1", "CA1", ]:
        df_fil = df[df['system']==system]
        x , y, yerr = df_fil["SetNum"], df_fil['Response','mean'], df_fil['Response','sem']
        
        ax.errorbar(x, y, yerr=yerr, **comp_format[system])

    ax.legend(loc=(0.02, 0.7))

    plt.savefig("doc_search.pdf", bbox_inches='tight', pad_inches=0.01)
    plt.savefig("doc_search.png", bbox_inches='tight', pad_inches=0.01)
    plt.show()


def plot_chem_search(data):
    fig, ax = plt.subplots()

    format = {
        'CA' : {
            'label':"CA-MS",
            'color':'#1f78b4',
            'fmt':'v-',
            # 'linestyle':'dashed', 
            # fmt='v'
        },
        'X' : {
            'label':"X-MS",
            'color':'#33a02c',
            'fmt':'^-',
            'alpha':0.5,
            'linewidth':5,
            # fmt='^-'
        },
        'CA-tr' : {
            'label':"CA-MS",
            'color':'#1f78b4',
            'fmt':'x',
            'linestyle':'dotted', 
        },
        'X-tr' : {
            'label':"X-MS",
            'color':'#33a02c',
            'fmt':'x',
            'linestyle':'dotted', 
        }
    }

    ax.set_yscale('log')
    ax.set_ylabel('Time (s)')


    plt.xscale('log')
    ax.set_xlabel('\#Chemical compounds')

    # xticks = [1000, 2000, 4000, 8000, 16000, 32000, 64000, 128000, 256000, 512000, 1000000, 2000000]
    # xlabels = [f"{x//1000}k" for x in xticks[:-2]] + ['1M', '2M']

    xticks = [1000, 2000, 8000, 32000, 128000, 512000, 2000000]
    xlabels = [f"{x//1000}k" for x in xticks[:-1]] + ['2M']
    plt.xticks(xticks, xlabels)

    yticks = [12,18,30,54, 102, 192, 378]
    # ylabels = [f"{y/1024}" for y in yticks]

    ax2 = ax.twinx()
    ax2.set_ylabel('Transfer cost (MB)')
    ax2.set_yscale('log')

    ax2.set_yticks(yticks)
    ax2.set_yticklabels(yticks)

    ax2.get_yaxis().set_minor_formatter(matplotlib.ticker.FuncFormatter(lambda y, _: ''))
    ax2.get_yaxis().set_tick_params(which='minor', size=0)
    ax2.get_yaxis().set_tick_params(which='minor', width=0) 

    df = data.groupby(['SetNum', 'system']).agg(['mean', 'sem'])
    df = df.reset_index()



    df_fil = df[df['system']=='X']
    x , y, yerr = df_fil["SetNum"], df_fil['Response','mean'], df_fil['Response','sem']
    ax.errorbar(x, y, yerr=yerr, **format['X'])
    x , y = df_fil["SetNum"], df_fil['RespSize','mean']
    y = [a/1024 + 6 for a in y] 
    print(y)
    ax2.errorbar(x, y, **format['X-tr'])


    df_fil = df[df['system']=='CA']
    x , y, yerr = df_fil["SetNum"], df_fil['Response','mean'], df_fil['Response','sem']
    ax.errorbar(x, y, yerr=yerr, **format['CA'])
    x , y = df_fil["SetNum"], df_fil['RespSize','mean']
    y = [a/1024 + 6 for a in y] 
    print(y)
    ax2.errorbar(x, y,  **format['CA-tr'])

    ax.legend(loc=(0.02, 0.7))
    # ax2.legend(loc=(0.685, 0.09))

    plt.savefig("chem_search.pdf", bbox_inches='tight', pad_inches=0.01)
    plt.savefig("chem_search.png", bbox_inches='tight', pad_inches=0.01)
    plt.show()


# %%
prefix = "benchmark_data/"
ca1 = read_bench(f'{prefix}chem_ca.json')
x1 = read_bench(f'{prefix}chem_xms.json')
ca1['system'] = 'CA'
x1['system'] = 'X'
data = pd.concat([x1, ca1])
data = data[['SetNum', 'Response', 'RespSize', 'system']]
# display(data) 
plot_chem_search(data)

# %%
benches = []
for system in ["X1", "X2", "CA1", "CA2"]:
    bench = read_bench(f'{prefix}doc_{system}.json')
    bench['system'] = system
    benches.append(bench)
data = pd.concat(benches)
data = data[['SetNum', 'Response', 'RespSize', 'system']]
data 
plot_doc_search(data)

# %%
# Plot small domain cardinality
def plot_sd_search(data):
    fig, ax = plt.subplots()

    ax.set_yscale('log')
    ax.set_ylabel('Time (ms)')

    plt.xscale('log')
    ax.set_xlabel('\#Sets')
    ticks = [1, 2, 4, 8, 16,  64,  256,  1024]
    plt.xticks(ticks, ticks)

    df = data.groupby(['SetNum', 'system']).agg(['mean', 'sem'])
    df = df.reset_index()

    comp_format = {
        'SD-256' : {
            'label':"Ours-256",
            'color':'#1f78b4',
            'fmt':'o-'
        },
        'SD-4k' : {
            'label':"Ours-4096",
            'color':'#a6cee3',
            'fmt':'v-'
        },

        'ruan_256' : {
            'label':"Ruan-256",
            'color':'#33a02c',
            'fmt':'*-'
        },
        'ruan_4k' : {
            'label':"Ruan-4096",
            'color':'#b2df8a',
            'fmt':'^-'
        }
    }

    for system in ["SD-256", "SD-4k", "ruan_256","ruan_4k"]:
        df_fil = df[df['system']==system]
        x , y, yerr = df_fil["SetNum"], df_fil['Cost','mean'], df_fil['Cost','sem']
        
        ax.errorbar(x, y, yerr=yerr, **comp_format[system])

    ax.legend(loc=(0.015, 0.435))

    plt.savefig("sd_ca_small_search.pdf", bbox_inches='tight', pad_inches=0.01)
    plt.show()


# %%
ruan_256 = [{'SetNum':ns, "Cost":30+2.7*ns} for ns in [2**i for i in range(13)]]
ruan_4k = [{'SetNum':ns, "Cost":295+14.9*ns} for ns in [2**i for i in range(13)]]
ruan_256 = pd.DataFrame(ruan_256)
ruan_4k = pd.DataFrame(ruan_4k)
ruan_256['system'] = "ruan_256"
ruan_4k['system'] = "ruan_4k"

# %%
sel = '7700_'
# sel = ""
sd_256 = read_bench(f'../bench/{sel}sm_ca_256_agg.json')
sd_4k = read_bench(f'../bench/{sel}sm_ca_4k_agg.json')
sd_256['system'] = "SD-256"
sd_4k['system'] = "SD-4k"
sd_256['Cost'] = 1000*(sd_256['Evaluation'] + sd_256['Query'] + sd_256['Response'])
sd_4k['Cost'] = 1000*(sd_4k['Evaluation'] + sd_4k['Query'] + sd_4k['Response'])

data = pd.concat([sd_256, sd_4k, ruan_256, ruan_4k])
data = data.query("SetNum < 2000")
plot_sd_search(data)
# %%
