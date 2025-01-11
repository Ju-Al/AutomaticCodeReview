from point import Point

class Line:
    def __init__(self, p1: Point, p2: Point):
        self.p1 = p1
        self.p2 = p2

    def get_slope(self) -> float:
        if self.p2.x == self.p1.x:
            raise ValueError("Slope is undefined for a vertical line.")
        return (self.p2.y - self.p1.y) / (self.p2.x - self.p1.x)

    def get_intercept(self) -> float:
        return self.p1.y - self.get_slope() * self.p1.x

    def is_on(self, point: Point) -> bool:
        # Check if a point lies on the infinite line
        slope = self.get_slope()
        return abs(point.y - (slope * point.x + self.get_intercept())) < 1e-9

    def get_p1(self) -> Point:
        return self.p1

    def get_p2(self) -> Point:
        return self.p2