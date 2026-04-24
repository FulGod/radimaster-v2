import { useEffect, useState } from 'react';
import { useAuth } from '../AuthContext';
import { api } from '../api';

interface SystemMetrics {
  system: {
    uptime_seconds: number;
    uptime_human: string;
    go_version: string;
    goroutines: number;
    cpu_cores: number;
  };
  memory: {
    alloc_mb: number;
    sys_mb: number;
    gc_cycles: number;
    gc_pause_total: string;
  };
  database: {
    status: string;
    open_connections: number;
    in_use: number;
    idle: number;
    max_open: number;
    wait_count: number;
    wait_duration: string;
  };
  websocket: {
    active_rooms: number;
    active_connections: number;
    redis_status: string;
  };
  data: {
    users: number;
    cases: number;
    classrooms: number;
  };
}

function StatusDot({ status }: { status: string }) {
  const isOk = status === 'connected';
  return (
    <span style={{
      display: 'inline-block',
      width: 10, height: 10,
      borderRadius: '50%',
      background: isOk ? '#22c55e' : '#ef4444',
      boxShadow: isOk ? '0 0 8px rgba(34,197,94,.5)' : '0 0 8px rgba(239,68,68,.5)',
      marginRight: 8,
    }} />
  );
}

function MetricCard({ title, children, accent = '#3b82f6' }: { title: string; children: React.ReactNode; accent?: string }) {
  return (
    <div style={{
      background: 'rgba(255,255,255,.04)',
      border: '1px solid rgba(255,255,255,.08)',
      borderRadius: 16,
      padding: '24px 28px',
      backdropFilter: 'blur(12px)',
      borderTop: `3px solid ${accent}`,
    }}>
      <div style={{ fontSize: 13, color: 'rgba(255,255,255,.5)', textTransform: 'uppercase', letterSpacing: 1, marginBottom: 16, fontWeight: 600 }}>
        {title}
      </div>
      {children}
    </div>
  );
}

function StatRow({ label, value, unit }: { label: string; value: string | number; unit?: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '8px 0', borderBottom: '1px solid rgba(255,255,255,.05)' }}>
      <span style={{ color: 'rgba(255,255,255,.6)', fontSize: 14 }}>{label}</span>
      <span style={{ color: '#fff', fontWeight: 600, fontSize: 15, fontFamily: "'JetBrains Mono', monospace" }}>
        {value}{unit && <span style={{ color: 'rgba(255,255,255,.4)', fontSize: 12, marginLeft: 4 }}>{unit}</span>}
      </span>
    </div>
  );
}

function BigNumber({ value, label, color = '#3b82f6' }: { value: number | string; label: string; color?: string }) {
  return (
    <div style={{ textAlign: 'center' }}>
      <div style={{ fontSize: 36, fontWeight: 700, color, fontFamily: "'JetBrains Mono', monospace", lineHeight: 1.2 }}>{value}</div>
      <div style={{ fontSize: 12, color: 'rgba(255,255,255,.5)', marginTop: 4, textTransform: 'uppercase', letterSpacing: 1 }}>{label}</div>
    </div>
  );
}

export default function DashboardPage() {
  const { token } = useAuth();
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [error, setError] = useState('');
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);

  const fetchMetrics = async () => {
    try {
      const res = await fetch(`${api.getApiBase()}/api/admin/metrics`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setMetrics(data);
      setLastUpdate(new Date());
      setError('');
    } catch (e) {
      setError((e as Error).message);
    }
  };

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(fetchMetrics, 5000); // refresh every 5s
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div style={{ minHeight: '100vh', background: 'linear-gradient(135deg, #0a0e1a 0%, #111827 50%, #0f172a 100%)', color: '#fff', padding: '32px 24px' }}>
      <div style={{ maxWidth: 1200, margin: '0 auto' }}>
        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 32 }}>
          <div>
            <h1 style={{ fontSize: 28, fontWeight: 700, margin: 0, display: 'flex', alignItems: 'center', gap: 12 }}>
              <span style={{ fontSize: 32 }}>📊</span> System Dashboard
            </h1>
            <p style={{ color: 'rgba(255,255,255,.5)', margin: '4px 0 0', fontSize: 14 }}>
              Real-time monitoring · RadiMaster V2
            </p>
          </div>
          <div style={{ textAlign: 'right' }}>
            {lastUpdate && (
              <div style={{ fontSize: 12, color: 'rgba(255,255,255,.4)' }}>
                Last update: {lastUpdate.toLocaleTimeString()} · Auto-refresh 5s
              </div>
            )}
            {error && <div style={{ color: '#ef4444', fontSize: 13, marginTop: 4 }}>⚠ {error}</div>}
          </div>
        </div>

        {!metrics ? (
          <div style={{ textAlign: 'center', padding: 80, color: 'rgba(255,255,255,.4)' }}>Loading metrics...</div>
        ) : (
          <>
            {/* Service Status Bar */}
            <div style={{
              display: 'flex', gap: 24, marginBottom: 32, padding: '16px 24px',
              background: 'rgba(255,255,255,.03)', borderRadius: 12, border: '1px solid rgba(255,255,255,.06)',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <StatusDot status={metrics.database.status} />
                <span style={{ fontSize: 14, color: 'rgba(255,255,255,.7)' }}>PostgreSQL</span>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <StatusDot status={metrics.websocket.redis_status} />
                <span style={{ fontSize: 14, color: 'rgba(255,255,255,.7)' }}>Redis PubSub</span>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <StatusDot status="connected" />
                <span style={{ fontSize: 14, color: 'rgba(255,255,255,.7)' }}>Go Backend</span>
              </div>
              <div style={{ marginLeft: 'auto', fontSize: 13, color: 'rgba(255,255,255,.4)' }}>
                Uptime: <strong style={{ color: '#22c55e' }}>{metrics.system.uptime_human}</strong>
              </div>
            </div>

            {/* Big Numbers */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 20, marginBottom: 32 }}>
              <MetricCard title="" accent="#3b82f6">
                <BigNumber value={metrics.system.goroutines} label="Goroutines" color="#3b82f6" />
              </MetricCard>
              <MetricCard title="" accent="#8b5cf6">
                <BigNumber value={metrics.websocket.active_connections} label="WS Connections" color="#8b5cf6" />
              </MetricCard>
              <MetricCard title="" accent="#22c55e">
                <BigNumber value={metrics.data.cases} label="Medical Cases" color="#22c55e" />
              </MetricCard>
              <MetricCard title="" accent="#f59e0b">
                <BigNumber value={`${metrics.memory.alloc_mb.toFixed(1)}`} label="Memory (MB)" color="#f59e0b" />
              </MetricCard>
            </div>

            {/* Detail Cards */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 20 }}>
              {/* System */}
              <MetricCard title="⚙️ System" accent="#3b82f6">
                <StatRow label="Go Version" value={metrics.system.go_version} />
                <StatRow label="CPU Cores" value={metrics.system.cpu_cores} />
                <StatRow label="Goroutines" value={metrics.system.goroutines} />
                <StatRow label="GC Cycles" value={metrics.memory.gc_cycles} />
                <StatRow label="GC Pause Total" value={metrics.memory.gc_pause_total} />
              </MetricCard>

              {/* Database */}
              <MetricCard title="🐘 PostgreSQL" accent="#22c55e">
                <StatRow label="Status" value={metrics.database.status === 'connected' ? '● Connected' : '○ Down'} />
                <StatRow label="Open Connections" value={metrics.database.open_connections} />
                <StatRow label="In Use" value={metrics.database.in_use} />
                <StatRow label="Idle" value={metrics.database.idle} />
                <StatRow label="Max Open" value={metrics.database.max_open} />
                <StatRow label="Wait Queue" value={metrics.database.wait_count} />
              </MetricCard>

              {/* WebSocket */}
              <MetricCard title="📡 WebSocket" accent="#8b5cf6">
                <StatRow label="Active Rooms" value={metrics.websocket.active_rooms} />
                <StatRow label="Live Connections" value={metrics.websocket.active_connections} />
                <StatRow label="Redis PubSub" value={metrics.websocket.redis_status === 'connected' ? '● Scaling' : '○ In-Memory'} />
                <div style={{ marginTop: 16, padding: '12px 16px', background: 'rgba(139,92,246,.1)', borderRadius: 8, fontSize: 12, color: 'rgba(255,255,255,.6)', lineHeight: 1.6 }}>
                  Redis PubSub enables horizontal scaling — WebSocket messages are relayed across all backend instances.
                </div>
              </MetricCard>
            </div>

            {/* Data Summary */}
            <div style={{ marginTop: 20 }}>
              <MetricCard title="📋 Data Summary" accent="#06b6d4">
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 24, padding: '8px 0' }}>
                  <BigNumber value={metrics.data.users} label="Users" color="#06b6d4" />
                  <BigNumber value={metrics.data.cases} label="Medical Cases" color="#22c55e" />
                  <BigNumber value={metrics.data.classrooms} label="Classrooms" color="#f59e0b" />
                </div>
              </MetricCard>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
