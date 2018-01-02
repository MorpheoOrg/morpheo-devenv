import sys
import os
import getopt
import numpy as np
import h5py
import json
import time
import hashlib


VALID_FILE_HASH = [
    "7182575d2f1fe035c0ce8cea70f93cd7",
    "86564dc69f8c9b081b5174ef562e1ac1",
    "f09058c9de0b55d482f8575f8e8e7628",
    "ca7a7b23a4d5cb655df97378e571f7c6",

    # Detargeted Value
    "de5a7197e3e61c7987d3a64731608187",
    "41b26abe9fd63ea31a8ea325bb9fb47c",
    "225cc8d69add816ca2506ffbf7d37615",
    "5a95ae04c2f7a6d92435bb0e84a1e170"
]


class Classifier():
    def __init__(self, volume):
        # Define directory and files where to fetch data and store output
        for v in ["train/pred", "test/pred", "model"]:
            directory = "%s/%s" % (volume, v)
            if not os.path.exists(directory):
                os.makedirs(directory)

        self.train_dir = os.path.join(volume, "train")
        self.test_dir = os.path.join(volume, "test")
        self.file_train_pred = '%s/train/pred/' % volume
        self.file_test_pred = '%s/test/pred/' % volume
        self.file_trained_model = '%s/model/model_trained.json' % volume
        self.model = []

    def predict(self, list_data, file_pred_dir):
        # check that model is present
        self.load_model(from_file=True)

        # loop over records, "predict" ;), and store predictions
        for f in list_data:
            X, _ = read_data([f])

            # prediction
            y_pred = np.random.rand(X.shape[0],)

            # Save predictions
            fname = f.split("/")[-1]
            file_pred = '%s%s' % (file_pred_dir, fname)
            print("Saving predictions under %s..." % file_pred)
            f = h5py.File(file_pred, "w")
            f.create_dataset("stages", data=y_pred)
            f.close()

    def load_model(self, from_file=False):
        try:
            with open(self.file_trained_model, 'r') as f:
                model = json.load(f)
        except FileNotFoundError:
            model = []
        if from_file and not model:
            print("[ERROR] Missing model_trained.json file")
            sys.exit(2)
        self.model = model

    def update_model(self, msg):
        info = {"timestamp": int(time.time()), "msg": msg}
        if self.model:
            info["id"] = self.model[-1]["id"] + 1
            self.model.append(info)
        else:
            info["id"] = 0
            self.model.append(info)
        with open(self.file_trained_model, 'w') as f:
            json.dump(self.model, f)


def read_data(list_data):
        X, y = [], []
        for fname in list_data:
            f = h5py.File(fname, "r")

            X.append(f["EEG1"][:])
            try:
                y.append(f["stages"][:])
            except KeyError:
                pass
            f.close()
        X = np.concatenate(X)
        if y:
            y = np.concatenate(y)
        return X, y


def check_data(list_data, data_type):
    if len(list_data) == 0:
        print("[ERROR] %s files are missing" % data_type)
        sys.exit(2)
    for fname in list_data:
        with open(fname, "rb") as f:
            h = hashlib.md5(f.read()).hexdigest()
            if h not in VALID_FILE_HASH:
                print("[ERROR] Invalid hash (%s) for file %s" % (h, fname))
                sys.exit(2)


def main(argv):
    try:
        opts, args = getopt.getopt(argv, "hV:T:", ["Volume=", "Task="])
    except getopt.GetoptError:
        print("[ERROR] Invalid arguments. "
              "Correct arguments: -V <volume> -T <train/predict>'")
        sys.exit(2)
    for opt, arg in opts:
        if opt == '-h':
            print('Arguments: -V <volume> -T <train/predict>')
            sys.exit()
        elif opt in ("-V", "--Volume"):
            volume = arg
        elif opt in ("-T", "--Task"):
            task_type = arg
    try:
        print("Starting task '%s' with volume '%s'..." % (task_type, volume))
    except NameError:
        print("[ERROR] Missing arguments (%s) to run python script. "
              "Correct arguments: -V <volume> -T <train/predict>" % opts)
        sys.exit(2)
    model = Classifier(volume)
    if task_type == "train":
        # Check data
        path = "{}/train/".format(volume)
        train_data = [path + f for f in os.listdir(path)
                      if os.path.isfile(os.path.join(path, f))]
        print("training data", train_data)
        check_data(train_data, "train")

        # simulate training
        model.load_model()
        model.update_model("train")

        # predicting
        model.predict(train_data, model.file_train_pred)

    if task_type in ["predict", "train"]:
        # Check data
        path = "{}/test/".format(volume)
        test_data = [path + f for f in os.listdir(path)
                     if os.path.isfile(os.path.join(path, f))]
        print("test data", test_data)
        check_data(test_data, "test")

        # predicting
        model.predict(test_data, model.file_test_pred)


if __name__ == "__main__":
    main(sys.argv[1:])
