import matplotlib.pyplot as plt
import numpy as np

def calculate_throughput(save):
    x_data = [3, 4, 5]
    raw_add = [1.017578215, 2.813627548, 5.366602778]
    y_add = [1000 * raw/50 for raw in raw_add]
    raw_mult = [1.902120465, 3.915597018, 11.60992823]
    y_mult = [1000 * raw/50 for raw in raw_mult]

    plt.plot(x_data, y_add, marker='o', label="MPC add")
    plt.plot(x_data, y_mult, marker='o', label="MPC mult")

    plt.xlim([2.5, 5.5])
    plt.xticks(x_data)
    plt.xlabel('Number of nodes')
    plt.ylabel('Average time per MPC (ms)')
    plt.title("Avg full MPC time")
    plt.legend(loc='upper left')
    plt.grid()

    if save is False:
        plt.show()
    else:
        plt.savefig('../results/full_mpc_time.pdf', bbox_inches='tight')
        plt.clf()

def mpc_details(save):
    # a+b+c
    label_add = 'PreMPC', 'MPC', 'PostMPC'
    raw_add = [4345, 25792, 13834]
    sum_add = sum(raw_add)
    size_add = [raw/sum_add for raw in raw_add]

    # (a+b)*c
    label_mix = 'PreMPC', 'MPC', 'PostMPC'
    raw_mix = [2641, 42654, 16513]
    sum_mix = sum(raw_mix)
    size_mix = [raw/sum_mix for raw in raw_mix]

    # a*b*c
    label_mult = 'PreMPC', 'MPC', 'PostMPC'
    raw_mult = [6195, 79189, 17072]
    sum_mult = sum(raw_mult)
    size_mult = [raw/sum_mult for raw in raw_mult]

   
    plt.pie(size_add, labels=label_add, startangle=180, autopct='%1.1f%%')
    plt.axis('equal')
    plt.title("MPC add")

    if save is False:
        plt.show()
    else:
        plt.savefig('../results/full_mpc_add_details.pdf', bbox_inches='tight')
        plt.clf()

    plt.pie(size_mult, labels=label_mult, startangle=180, autopct='%1.1f%%')
    plt.axis('equal')
    plt.title("MPC mult")

    if save is False:
        plt.show()
    else:
        plt.savefig('../results/full_mpc_mult_details.pdf', bbox_inches='tight')
        plt.clf()

    plt.pie(size_mix, labels=label_mix, startangle=180, autopct='%1.1f%%')
    plt.axis('equal')
    plt.title("MPC mix")

    if save is False:
        plt.show()
    else:
        plt.savefig('../results/full_mpc_mix_details.pdf', bbox_inches='tight')
        plt.clf()
    