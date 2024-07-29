	}
	return ctx, nil
}

func isWorkloadAPIMethod(fullMethod string) bool {
	return strings.HasPrefix(fullMethod, workloadAPIMethodPrefix)
}

func hasSecurityHeader(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	return ok && len(md["workload.spiffe.io"]) == 1 && md["workload.spiffe.io"][0] == "true"
}
