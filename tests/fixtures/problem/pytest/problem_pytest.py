import os
import sys
import glob
import json
import getopt
import shutil
import numpy as np
import h5py
from pprint import pprint


def remove_target(dir_test_files, dir_new_test_files, target_name):
    """
    copy test file in dir_new_test_file and remove target

    :param dir_test_files: path where test data are (test_xxx)
    :param dir_new_test_files: path of resulting test data files without target
    :param target_name: target column name
    :type dir_test_files: string
    :type dir_new_test_files: string
    :type target_name: string
    """
    print("Removing targets from {} into {}...".format(
        dir_test_files, dir_new_test_files
    ))
    # Get a list of all train and test files in volume dir_test_file
    files = [f for f in glob.glob('{}/*'.format(dir_test_files)) if
             os.path.isfile(f)]
    print("Copying and detargeting test files: {}".format(files))
    for t_file in files:
        # Copy file in dir_new_test_file
        file_name = t_file.split('/')[-1]
        new_test_file = '{}{}'.format(dir_new_test_files, file_name)
        shutil.copy(t_file, new_test_file)
        # Remove target and save file
        with h5py.File(new_test_file, "a") as f:
            f.__delitem__(target_name)


def compute_perf_files(type_files, dir_true_files, dir_pred_files, metric,
                       target_name):
    """
    compute performances given a metric for a train or test files

    :param type_files: prefix of true data files (train or test)
    :param dir_test_files: path where are true data (<type_files>_<uuid>)
    :param dir_pred_files: path where are predicted data\
        (pred_<type_files>_<uuid>)
    :param metric: performance metrics
    :param target_name: target column name
    :type dir_test_files: string
    :type dir_pred_files: string
    :type metric: function(y_true, y_pred)
    :type target_name: string
    :return: returns a dictionary with performances estimated on all\
        <type_files> files (<type_files> key), and on each file\
        (<type_files>_uuid keys)
    """
    # Get a list of all <type_files> files
    true_files = [f for f in glob.glob('{}/*'.format(dir_true_files))
                  if os.path.isfile(f)]
    # Put true values in array and corresponding predicted values in another
    true_label = []
    pred_label = []
    # Put perf for each file in dictionary
    perf = {}
    perf['%s_perf' % type_files] = {}
    for t_file in true_files:
        # get uuid
        uuid_file = os.path.basename(t_file)
        # true values
        f_true = h5py.File(t_file, "r")
        true_label_file = list(f_true[target_name][:])
        true_label.extend(true_label_file)
        f_true.close()

        # corresponding predicted values
        file_name = os.path.basename(t_file)
        path = os.path.join(dir_pred_files, file_name)
        f_pred = h5py.File(path, "r")
        pred_label_file = list(f_pred[target_name][:])
        pred_label.extend(pred_label_file)
        f_pred.close()

        # compute performance for this file
        perf['%s_perf' % type_files][uuid_file] = metric(true_label_file,
                                                         pred_label_file)
    # Compute performance on test data
    perf[type_files] = metric(true_label, pred_label)
    return perf


def compute_perf(dir_true_test_files, dir_true_train_files,
                 dir_pred_test_files, dir_pred_train_files,
                 metric, target_name):
    """
    compute performances for train and test files, and each file separately

    :param dir_true_test_files: path where are true test data
    :param dir_true_train_files: path where are true train data
    :param dir_pred_test_files: path where are predicted test data
    :param dir_pred_train_files: path where are predicted train data
    :param metric: performance metrics
    :param target_name: target column name
    :type dir_test_files: string
    :type dir_pred_files: string
    :type metric: function(y_true, y_pred)
    :type target_name: string
    :return: returns a dictionary with performances estimated on all train\
        files (train key), all test files (test key), and on each file\
        (train/test_uuid keys)
    """
    # Compute performances on train files
    perf_train = compute_perf_files('train', dir_true_train_files,
                                    dir_pred_train_files, metric, target_name)
    # Compute performances on test files
    perf_test = compute_perf_files('test', dir_true_test_files,
                                   dir_pred_test_files, metric, target_name)
    # Merge two dictionary
    perf = {**perf_train, **perf_test}
    perf["perf"] = perf.pop("test")
    del perf["train"]
    print("perf computed:")
    pprint(perf)
    return perf


def metric(y1, y2):
    return np.mean(np.absolute(np.array(y2) - np.array(y1)))


def main(argv):
    target_name = "stages"
    file_perf = "performance.json"
    try:
        opts, args = getopt.getopt(argv, "hT:i:s:",
                                   ["Task", "hiddenpath=", "submissionpath="])
    except getopt.GetoptError:
        print('[ERROR] Invalid arguments. '
              'Correct arguments: -T <detarget/perf> '
              '-i <hidden_path> -s <submission_path>')
        sys.exit(2)
    for opt, arg in opts:
        if opt == '-h':
            print('Arguments: -T <detarget/perf> '
                  '-i <hidden_path> -s <submission_path>')
            sys.exit()
        elif opt in ("-i", "--hiddenpath"):
            hidden_path = arg
        elif opt in ("-s", "--submissionpath"):
            submission_path = arg
        elif opt in ("-T", "--Task"):
            task_type = arg
    try:
        print("Starting task '%s' with hidden_path '%s' and submission_path"
              " '%s'..." % (task_type, hidden_path, submission_path))
    except NameError:
        print("[ERROR] Missing arguments (%s) to run python script. "
              "Correct arguments: -T <detarget/perf> "
              "-i <hidden_path> -s <submission_path>" % opts)
        sys.exit(2)
    dir_true_test_files = hidden_path + "/test/"
    dir_perf_file = hidden_path + "/perf/"
    dir_detargeted_test_files = submission_path + "/test/"
    dir_true_train_files = submission_path + "/train/"
    dir_pred_test_files = submission_path + "/test/pred"
    dir_pred_train_files = submission_path + "/train/pred"

    if task_type == "detarget":
        # create dir_detargeted_test_files if it does not exist
        if not os.path.exists(dir_detargeted_test_files):
            os.makedirs(dir_detargeted_test_files)
        remove_target(dir_true_test_files, dir_detargeted_test_files,
                      target_name)
    elif task_type == "perf":
        perf = compute_perf(dir_true_test_files, dir_true_train_files,
                            dir_pred_test_files, dir_pred_train_files, metric,
                            target_name)
        if not os.path.exists(dir_perf_file):
            os.makedirs(dir_perf_file)
        with open(os.path.join(dir_perf_file, file_perf), 'w') as f:
            json.dump(perf, f)


if __name__ == "__main__":
    main(sys.argv[1:])
