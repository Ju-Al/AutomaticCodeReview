package com.getcapacitor.annotation;
     * The Android string for the permission.
     * Eg: Manifest.permission.ACCESS_COARSE_LOCATION
     *     or "android.permission.ACCESS_COARSE_LOCATION"
    String permission() default "";

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;

/**
 * Permission annotation for use with @CapacitorPlugin
 */
@Retention(RetentionPolicy.RUNTIME)
public @interface Permission {
    /**
     * An array of Android permission strings.
     * Eg: {Manifest.permission.ACCESS_COARSE_LOCATION}
     *     or {"android.permission.ACCESS_COARSE_LOCATION"}
     */
    String[] permission() default {};

    /**
     * An optional name to use instead of the Android permission string.
     */
    String alias() default "";
}