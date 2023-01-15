import matplotlib.pyplot as plt
import numpy as np

def simple_txn_throughput(save):
    x_data = [2, 4, 6, 8, 10, 12, 14, 16, 18, 20]
    raw_data = [510.143212, 1197.252827, 1600.888948, 
                4145.237745, 8074.317746, 12548.058258, 
                17215.797999, 21148.166825, 25463.041297, 
                32479.116068]
    y_data = [raw/1000 for raw in raw_data]

    plt.plot(x_data, y_data, marker='o', label="Avg transaction commit time in blockchain")
    plt.xlim([0, 22])
    plt.xticks(x_data)
    # plt.ylim(0)
    plt.xlabel('Number of nodes')
    plt.ylabel('Avg time per transaction (ms)')
    plt.title("Avg transaction commit time in blockchain")
    plt.grid()

    if save is False:
        plt.show()
    else:
        plt.savefig('../results/bc_txn_time.pdf', bbox_inches='tight')
        plt.clf()