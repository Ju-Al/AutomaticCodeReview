
        public static void ExecuteQuery(string connectionString, string queryString, int version)
        {
            using (var connection = new SqlConnection(connectionString))
            {
                connection.Open();
                SqlCommand command = connection.CreateCommand();
                command.CommandText = queryString;
                command.CommandType = CommandType.Text;

                try
                {
                    command.ExecuteNonQueryAsync();
                }
                catch (SqlException)
                {
                    ExecuteUpsertQuery(connectionString, version, "failed");
                    throw;
                }
            }
        }

        public static void ExecuteUpsertQuery(string connectionString, int version, string status)
        {
            using (var connection = new SqlConnection(connectionString))
            {
                var upsertCommand = new SqlCommand("dbo.UpsertSchemaVersion", connection)
                {
                    CommandType = CommandType.StoredProcedure,
                };

                upsertCommand.Parameters.AddWithValue("@version", version);
                upsertCommand.Parameters.AddWithValue("@status", status);

                connection.Open();
                upsertCommand.ExecuteNonQuery();
            }
        }
    }
}
