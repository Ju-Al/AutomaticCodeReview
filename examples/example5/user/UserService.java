package example5.user;

public class UserService {
    private MySQLDatabase database;

    public UserService() {
        this.database = new MySQLDatabase();
    }

    public String getUser(int id) {
        return database.getUserData(id);
    }

    public void addUser(int id, String data) {
        this.database.addUserData(id, data);
    }

    public void updateUser(int id, String data) {
        this.database.updateUserData(id, data);
    }

    public void deleteUser(int id) {
        this.database.deleteUserData(id);
    }
}