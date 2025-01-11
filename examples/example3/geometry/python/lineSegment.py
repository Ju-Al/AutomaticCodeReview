from line import Line
from point import Point
from math import sqrt

class LineSegment(Line):
    def __init__(self, p1: Point, p2: Point):
        super().__init__(p1, p2)

    def get_length(self) -> float:
        dx = self.p2.x - self.p1.x
        dy = self.p2.y - self.p1.y
        return sqrt(dx * dx + dy * dy)

    def is_on(self, point: Point) -> bool:
        # Check if the point lies within the segment
        if not super().is_on(point):
            return False
        min_x, max_x = sorted([self.p1.x, self.p2.x])
        min_y, max_y = sorted([self.p1.y, self.p2.y])
        return min_x <= point.x <= max_x and min_y <= point.y <= max_y