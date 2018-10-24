db.createUser(
    {
      user: "user1",
      pwd: "example",
      roles: ["readWrite"]
    }
);

rs.initiate();