import matplotlib.pyplot as plt
import numpy as np

def simple_txn_throughput(save):
    raw_sent = [[1261, 1268, 1268, 1264, 1261, 1271, 1261, 1273]]
    raw_commit = [[1260, 1260, 1260, 1260, 1260, 1270, 1260, 1270]]
    y_sent = np.array(raw_sent)
    y_commit = np.array(raw_commit)
    x_data = [3, 5]

    plt.bar(x_data, y_sent, label="Total Txn Sent")
    plt.bar(x_data, y_commit, label="Total Txn Committed")

    plt.xlim([0, 100])
    plt.ylim(0)
    plt.xlabel('Percentage of Data Loss')
    plt.ylabel('Average Download Parity Blocks Ratio')
    plt.title("Part of Data Missing and All Parity Present")
    if save is False:
        plt.show()
    else:
        plt.savefig('results/only_data_download_overhead.png', bbox_inches='tight')
        plt.clf()