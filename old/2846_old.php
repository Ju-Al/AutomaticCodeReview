        });

        Schema::table('action_logs', function ($table) {
            $table->dropColumn('accept_signature');
        });
    }
}
