import base64
import unittest
import project


class ProjectContractTests(unittest.TestCase):
    def fixture(self):
        return {'profile':'python-unittest-v1','directory':'.','files':{
            'calc.py':base64.b64encode(b'def add(a,b): return a-b\n').decode(),
            'test_calc.py':base64.b64encode(b'import unittest\n').decode()},
            'edits':[{'file':'calc.py','old':'a-b','new':'a+b'}]}

    def test_preserves_acceptance_tests(self):
        _,_,original,candidate,_=project.prepare(self.fixture())
        self.assertEqual(original['test_calc.py'],candidate['test_calc.py'])
        self.assertNotEqual(project.manifest_hash(original),project.manifest_hash(candidate))

    def test_blocks_test_replacement(self):
        data=self.fixture();data['edits']=[{'file':'test_calc.py','old':'import unittest','new':'pass'}]
        with self.assertRaises(ValueError):project.prepare(data)

    def test_blocks_paths(self):
        for name in ['../escape','/etc/passwd','a/../b','.git/config','a\\b','a//b']:
            with self.subTest(name=name),self.assertRaises(ValueError):project.safe_path(name)

    def test_requires_tests(self):
        data=self.fixture();del data['files']['test_calc.py']
        with self.assertRaises(ValueError):project.prepare(data)

    def test_manifest_order_stable(self):
        self.assertEqual(project.manifest_hash({'a':b'1','b':b'2'}),project.manifest_hash({'b':b'2','a':b'1'}))

if __name__=='__main__':unittest.main()
