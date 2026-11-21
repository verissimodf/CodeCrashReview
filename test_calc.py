import unittest
from calc import add
class TestAdd(unittest.TestCase):
    def test_add(self): self.assertEqual(add(2, 3), 5)
    def test_negative(self): self.assertEqual(add(-2, -3), -5)
