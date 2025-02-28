class MySQLDatabase:
    def __init__(self):
        self.user_data = {
            1: "Alice",
            2: "Bob",
            3: "Charlie"
        }
    
    def get_user_data(self, id: int) -> str:
        if id in self.user_data:
            return f"User ID: {id}, Name: {self.user_data[id]}"
        return "User not found"
    
    def add_user_data(self, id: int, name: str) -> None:
        if id in self.user_data:
            print("User ID already exists. Use update_user_data to modify.")
        else:
            self.user_data[id] = name
            print("User added successfully.")
    
    def update_user_data(self, id: int, name: str) -> None:
        if id in self.user_data:
            self.user_data[id] = name
            print("User updated successfully.")
        else:
            print("User ID does not exist. Use add_user_data to add new user.")
    
    def delete_user_data(self, id: int) -> None:
        if id in self.user_data:
            del self.user_data[id]
            print("User deleted successfully.")
        else:
            print("User ID not found.")

class UserService:
    def __init__(self):
        self.database = MySQLDatabase()
    
    def get_user(self, id: int) -> str:
        return self.database.get_user_data(id)
    
    def add_user(self, id: int, data: str) -> None:
        self.database.add_user_data(id, data)
    
    def update_user(self, id: int, data: str) -> None:
        self.database.update_user_data(id, data)
    
    def delete_user(self, id: int) -> None:
        self.database.delete_user_data(id)

# Example usage (main)
if __name__ == "__main__":
    user_service = UserService()
    user_service.add_user(1, "Daiana")
    user_service.add_user(2, "Tom")
    user_service.update_user(2, "Tom Taylor")
    user_service.delete_user(1)