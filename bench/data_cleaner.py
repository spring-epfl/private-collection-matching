import json
import pandas as pd

# Record doc:
#  name: str
#  SetNum: int
#  Latency: seconds (float)
#  client_comp: seconds (float)
#  server_comp: seconds (float)
#  communication: bytes (int)

def parse_bashtime_fields(df: pd.DataFrame):
    df['Latency'] =  df['real_client'] 
    df['client_comp'] =  df['user_client'] +df['sys_client']
    df['server_comp'] =  df['user_server'] +df['sys_server']

def parse_circuit(client_log_addr: str, serv_log_addr: str, name: str) -> pd.DataFrame:
    clt = pd.read_csv(client_log_addr)
    serv = pd.read_csv(serv_log_addr)
    df = pd.merge(clt, serv, left_index=True, right_index=True, suffixes=('_client', '_server'))
    
    df['name'] = name
    df['SetNum'] =  df['SetNum_client'] 
    df['communication'] =  (1157397+655492)*df['SetNum'] 
    parse_bashtime_fields(df)

    return df

def parse_emp(client_csv_addr: str, serv_csv_addr: str, name: str) -> pd.DataFrame:
    clt = pd.read_csv(client_csv_addr)
    serv = pd.read_csv(serv_csv_addr)
    df = pd.merge(clt, serv, left_index=True, right_index=True, suffixes=('_client', '_server'))
    
    df['name'] = name
    df['SetNum'] =  df['SetNum_client'] 
    df['communication'] =  df['com_client'] + df['com_server']
    parse_bashtime_fields(df)

    return df

def add_gopsi_derived_fields(df):
    df['client_comp'] = df["Query"]+df["QueryMarshal"]+df["Evaluation"]
    df['server_comp'] = df["Response"]+df["RespMarshal"]
    df['communication'] = df["RespSize"]+df["QuerySize"]


def parse_gopsi(file_addr, name=''):
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
    df['name'] = name
    add_gopsi_derived_fields(df)
    return df

def extrapolate_spot(dir_addr:str, name:str) -> pd.DataFrame:
    clt = pd.read_csv(dir_addr+'spot_client.log')
    serv = pd.read_csv(dir_addr+'spot_server.log')
    raw = pd.merge(clt, serv, left_index=True, right_index=True, suffixes=('_client', '_server'))
    
    raw['SetNum'] =  raw['SetNum_client'] 
    raw['communication'] =  (30534)*raw['SetNum'] 
    parse_bashtime_fields(raw)

    base = raw.query('SetNum == 8')
    dfs = []
    for ext in [2**i for i in range(10)]:
        dfs.append(base.multiply(ext))
    df = pd.concat(dfs, ignore_index=True)
    df['name'] = name

    return df



def get_gopsi_bench(dir_addr) -> list[pd.DataFrame]:
    ca = parse_gopsi(dir_addr+'doc_ca_ms.json', 'CA-MS')
    x = parse_gopsi(dir_addr+'doc_x_ms.json', 'X-MS')
    return [ca, x]


def create_individual_dfs(raw_dir_addr: str, out_dir_addr: str, enable_spot = False):
    circuit = parse_circuit(raw_dir_addr+'circuit-time-client.csv', raw_dir_addr+'circuit-time-server.csv', 'Circuit-PSI')
    circuit.to_csv(out_dir_addr+'circuit_psi.csv')
    circuit.to_pickle(out_dir_addr+'circuit_psi.pkl')

    emp_ca = parse_emp(raw_dir_addr+'emp_ca_client.csv', raw_dir_addr+'emp_ca_server.csv', 'EMP-CA')
    emp_ca.to_csv(out_dir_addr+'emp_ca.csv')
    emp_ca.to_pickle(out_dir_addr+'emp_ca.pkl')

    emp_x = parse_emp(raw_dir_addr+'emp_x_client.csv', raw_dir_addr+'emp_x_server.csv', 'EMP-X')
    emp_x.to_csv(out_dir_addr+'emp_x.csv')
    emp_x.to_pickle(out_dir_addr+'emp_x.pkl')

    if enable_spot:
        spot = extrapolate_spot(raw_dir_addr, 'SpOT')
        spot.to_csv(out_dir_addr+'spot.csv')
        spot.to_pickle(out_dir_addr+'spot.pkl')


def load_full_df(data_dir_addr: str, approach_names: list[str]):
    # approach_names in ['circuit_psi', 'emp_x', 'emp_ca', 'spot']
    base_addr = data_dir_addr+'agg/'
    bench_columns = ['name','SetNum','Latency','client_comp','server_comp','communication',]

    dfs = get_gopsi_bench(base_addr)
    for ap_name in approach_names:
        dfs.append(pd.read_pickle(f'{base_addr}{ap_name}.pkl'))

    full_df = pd.concat(dfs)
    full_df = full_df[bench_columns]
    full_df['com_MiB'] = full_df['communication'] / (1024*1024)
    return full_df



def read_json_bench(file_addr):
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
