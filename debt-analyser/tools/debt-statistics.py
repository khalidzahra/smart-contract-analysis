import csv
import numpy as np
import matplotlib.pyplot as plt

def create_cdf_from_dict(data_dict, ylabel="CDF", xlabel="Value", title="Cumulative Distribution Function (CDF)", save=None):
    data_list = []
    for key, value in data_dict.items():
        if key > 100:
            continue
        data_list.extend([key] * value)

    # Sort the data
    data_list.sort()

    # Compute the CDF
    data_array = np.array(data_list)
    cdf = np.arange(1, len(data_array) + 1) / len(data_array)

    # Plot the CDF
    plt.plot(data_array, cdf, marker='.', linestyle='none')
    plt.xlabel(xlabel)
    plt.ylabel(ylabel)
    plt.title(title)
    plt.grid(True)
    if save == None:
        plt.show()
    else:
        plt.savefig(save)

def compute_statistics(csv_file_path):
    dataset_size = 0
    anomalous_contracts = []
    version_count = dict()
    total_initial_debt = 0
    
    with open(csv_file_path, newline='') as csvfile:
        reader = csv.reader(csvfile)

        for row in reader:
            # Invalid format
            if row[0][0] != '0':
                continue

            # ========== versions ==========
            versions = len(row) - 2
            if versions > 100:
                anomalous_contracts.append(row[0] + " " + row[1])
            else:
                dataset_size += 1
                 # ========== debt ==========
                total_initial_debt += int(row[2]) # third column contains the first version
            if versions in version_count:
                version_count[versions] += 1
            else:
                version_count[versions] = 1

           
    
    print("DATASET SIZE:", dataset_size)
    print('\n\n')
    print("CONTRACT VERSIONS:\n", '\n'.join(f"{key}: {version_count[key]}" for key in sorted(version_count.keys())))
    print('\n\n')
    print("ANOMALOUS CONTRACTS:\n", '\n'.join(anomalous_contracts))
    print('\n\n')
    print("Average Initial Debt:\n", total_initial_debt / dataset_size)
    print('\n\n')
    sorted_keys = sorted(version_count.keys())
    print("Median Initial Debt:\n", sorted_keys[(len(sorted_keys) - 1) // 2])

    create_cdf_from_dict(version_count, xlabel="Number of Versions", title="", save="version_num_cdf")



csv_file_path = '../out_data/total_debt.csv'
compute_statistics(csv_file_path)