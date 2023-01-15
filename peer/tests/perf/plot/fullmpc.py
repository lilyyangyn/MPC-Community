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