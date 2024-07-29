
        MainApp.bus().register(this);
        super.onStart();
        loadFromSP();
        loopHandler.postDelayed(refreshLoop, T.mins(1).msecs());
    }

    @Override
    protected void onStop() {
        loopHandler.removeCallbacks(refreshLoop);
        Context context = MainApp.instance().getApplicationContext();
        context.stopService(new Intent(context, LocationService.class));

        MainApp.bus().unregister(this);
    }

    public List<AutomationEvent> getAutomationEvents() {
        return automationEvents;
    }

    public EventLocationChange getEventLocationChange() {
        return eventLocationChange;
    }

    public EventChargingState getEventChargingState() {
        return eventChargingState;
    }

    public EventNetworkChange getEventNetworkChange() {
        return eventNetworkChange;
    }

    private void storeToSP() {
        JSONArray array = new JSONArray();
        try {
            for (AutomationEvent event : getAutomationEvents()) {
                array.put(new JSONObject(event.toJSON()));
            }
        } catch (JSONException e) {
            e.printStackTrace();
        }
        SP.putString(key_AUTOMATION_EVENTS, array.toString());
    }

    private void loadFromSP() {
        automationEvents.clear();
        String data = SP.getString(key_AUTOMATION_EVENTS, "");
        if (!data.equals("")) {
            try {
                JSONArray array = new JSONArray(data);
                for (int i = 0; i < array.length(); i++) {
                    JSONObject o = array.getJSONObject(i);
                    AutomationEvent event = new AutomationEvent().fromJSON(o.toString());
                    automationEvents.add(event);
                }
            } catch (JSONException e) {
                e.printStackTrace();
            }
        }

    }

    @Subscribe
    public void onEventPreferenceChange(EventPreferenceChange e) {
        if (e.isChanged(R.string.key_location)) {
            Context context = MainApp.instance().getApplicationContext();
            context.stopService(new Intent(context, LocationService.class));
            context.startService(new Intent(context, LocationService.class));
        }
    }

    @Subscribe
    public void onEvent(EventAutomationDataChanged e) {
        storeToSP();
    }

    @Subscribe
    public void onEventLocationChange(EventLocationChange e) {
        eventLocationChange = e;
        if (e != null)
            log.debug(String.format("Grabbed location: %f %f Provider: %s", e.location.getLatitude(), e.location.getLongitude(), e.location.getProvider()));

        processActions();
    }

    @Subscribe
    public void onEventChargingState(EventChargingState e) {
        eventChargingState = e;
        processActions();
    }

    @Subscribe
    public void onEventNetworkChange(EventNetworkChange e) {
        eventNetworkChange = e;
        processActions();
    }

    @Subscribe
    public void onEventAutosensCalculationFinished(EventAutosensCalculationFinished e) {
        processActions();
    }

    private synchronized void processActions() {
        if (!isEnabled(PluginType.GENERAL))
            return;

        if (L.isEnabled(L.AUTOMATION))
            log.debug("processActions");
        for (AutomationEvent event : getAutomationEvents()) {
            if (event.getTrigger().shouldRun()) {
                List<Action> actions = event.getActions();
                for (Action action : actions) {
                    action.doAction(new Callback() {
                        @Override
                        public void run() {
                            StringBuilder sb = new StringBuilder();
                            sb.append(DateUtil.timeString(DateUtil.now()));
                            sb.append(" ");
                            sb.append(result.success ? "☺" : "X");
                            sb.append(" ");
                            sb.append(event.getTitle());
                            sb.append(": ");
                            sb.append(action.shortDescription());
                            sb.append(": ");
                            sb.append(result.comment);
                            executionLog.add(sb.toString());
                            if (L.isEnabled(L.AUTOMATION))
                                log.debug("Executed: " + sb.toString());
                            MainApp.bus().post(new EventAutomationUpdateGui());
                        }
                    });
                }
                event.getTrigger().executed(DateUtil.now());
            }
        }
        storeToSP(); // save last run time
    }
}
