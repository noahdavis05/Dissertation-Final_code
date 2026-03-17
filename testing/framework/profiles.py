"""
A class which just stores variables. These variables are profiles of the test,
e.g. what will happen in a test.
"""
class TestProfile:

    def __init__(self):
        self.test_type = "stress-ng" # tests can be stress ng or teastore, but will just implement stress-ng to begin
        self.num_pods = 50
        self.manifest_path = "" # path which a manifest belongs for this test

