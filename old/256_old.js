    const result = await rollup({
        input: inputPath,
        plugins: [
            replace({
                __VERSION__: version
            })
        ]
    });

    await result.write({
        file: outputPath,
        format: "cjs",
        sourcemap: false, // keep typescript generated map to stay consistent with the rest of the files.
    });

    console.log("Compiler version: ", version);
}

if (!version || typeof version !== "string") {
    throw new Error(
        "Failed to update compiler version. Expected version value as a string, received: " +
            version
    );
}

updateVersion(version);