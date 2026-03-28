import { useState, useEffect, useCallback, useRef } from "react";
import type { DashboardOverview, ServicesData, GitInfo, ProcessInfo } from "../types";
import { fetchOverview, fetchServices, fetchGitLog, fetchProcesses } from "../api/client";

export function useDashboard() {
  const [overview, setOverview] = useState<DashboardOverview | null>(null);
  const [services, setServices] = useState<ServicesData | null>(null);
  const [gitInfo, setGitInfo] = useState<GitInfo | null>(null);
  const [processes, setProcesses] = useState<ProcessInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [paused, setPaused] = useState(false);
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());
  const [gitPath, setGitPath] = useState("");

  const pausedRef = useRef(paused);
  useEffect(() => { pausedRef.current = paused; }, [paused]);

  const gitPathRef = useRef(gitPath);
  useEffect(() => { gitPathRef.current = gitPath; }, [gitPath]);

  const initialLoadDone = useRef(false);

  const loadOverview = useCallback(async () => {
    try {
      const data = await fetchOverview();
      setOverview(data);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load overview");
    }
  }, []);

  const loadServices = useCallback(async () => {
    try {
      const data = await fetchServices();
      setServices(data);
    } catch { /* partial failure ok */ }
  }, []);

  const loadProcesses = useCallback(async () => {
    try {
      const data = await fetchProcesses();
      setProcesses(data);
    } catch { /* partial failure ok */ }
  }, []);

  const loadGit = useCallback(async (path: string) => {
    if (!path) return;
    try {
      const data = await fetchGitLog(path);
      setGitInfo(data);
    } catch { /* partial failure ok */ }
  }, []);

  // Initial load + polling
  useEffect(() => {
    const doLoad = async () => {
      if (initialLoadDone.current) return;
      initialLoadDone.current = true;
      setLoading(true);
      await Promise.all([loadOverview(), loadServices(), loadProcesses()]);
      setLastRefresh(new Date());
      setLoading(false);
    };
    doLoad();

    const overviewTimer = setInterval(() => {
      if (!pausedRef.current) loadOverview();
    }, 10000);
    const servicesTimer = setInterval(() => {
      if (!pausedRef.current) loadServices();
    }, 30000);
    const processTimer = setInterval(() => {
      if (!pausedRef.current) loadProcesses();
    }, 15000);

    return () => {
      clearInterval(overviewTimer);
      clearInterval(servicesTimer);
      clearInterval(processTimer);
    };
  }, [loadOverview, loadServices, loadProcesses]);

  const refresh = useCallback(async () => {
    setLoading(true);
    await Promise.all([loadOverview(), loadServices(), loadProcesses()]);
    if (gitPathRef.current) await loadGit(gitPathRef.current);
    setLastRefresh(new Date());
    setLoading(false);
  }, [loadOverview, loadServices, loadProcesses, loadGit]);

  const updateGitPath = useCallback((path: string) => {
    setGitPath(path);
    loadGit(path);
  }, [loadGit]);

  return {
    overview, services, gitInfo, processes,
    loading, error, lastRefresh,
    paused, setPaused,
    gitPath, updateGitPath,
    refresh,
  };
}
