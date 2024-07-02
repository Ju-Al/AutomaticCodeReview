class Garden(object):

    __plant_names = {"C": "Clover", "G": "Grass",
                     "R": "Radishes", "V": "Violets"}
    STUDENTS = [
        "Alice",
        "Bob",
        "Charlie",
        "David",
        "Eve",
        "Fred",
        "Ginny",
        "Harriet",
        "Ileana",
        "Joseph",
        "Kincaid",
        "Larry",
    ]
    PLANTS = {"C": "Clover", "G": "Grass", "R": "Radishes", "V": "Violets"}

    def __init__(self, diagram, students=None):
        students = sorted(students or self.STUDENTS)
        front, back = diagram.splitlines()
        self.cups = {}
        for idx, student in enumerate(students[: len(front)]):
            start = idx * 2
            stop = start + 2
            self.cups.setdefault(student, [])
            self.cups[student].extend(self.PLANTS[p] for p in front[start:stop])
            self.cups[student].extend(self.PLANTS[p] for p in back[start:stop])

    def plants(self, student):
        return self.cups.get(student, [])