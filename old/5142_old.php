            $table->boolean('login_common_disabled')->nullable()->default(0);
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::table('settings', function (Blueprint $table) {
            $table->dropColumn('login_remote_user_enabled');
            $table->dropColumn('login_common_disabled');
        });
    }
}
