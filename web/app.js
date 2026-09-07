const state = { events: [], filter: 'all', selected: null };
const timeline = document.querySelector('#timeline');
const emptyState = document.querySelector('#empty-state');
const connection = document.querySelector('#connection');

function safe(value) {
  return String(value ?? '').replace(/[&<>'"]/g, (char) => ({
    '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;'
  }[char]));
}

function relativeTime(iso) {
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(iso).getTime()) / 1000));
  if (seconds < 10) return 'now';
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
  return `${Math.floor(seconds / 3600)}h`;
}

function render() {
  const visible = state.events
    .filter((event) => state.filter === 'all' || event.kind === state.filter)
    .slice().reverse();
  emptyState.classList.toggle('hidden', visible.length > 0);
  timeline.innerHTML = visible.map((event) => `
    <li class="timeline-item ${safe(event.kind)} ${state.selected === event.sequence ? 'selected' : ''}" data-sequence="${event.sequence}" tabindex="0">
      <time datetime="${safe(event.time)}">${relativeTime(event.time)}</time>
      <span class="rail"><span class="dot"></span></span>
      <article class="timeline-card">
        <header><h4>${safe(event.title)}</h4><span class="tag ${safe(event.risk)}">${safe(event.risk)}</span></header>
        <p>${safe(event.summary)}</p>
        <div class="event-tags">
          <span class="tag actor">${safe(event.actor)}</span>
          <span class="tag">${safe(event.confidence)}</span>
          ${event.path ? `<span class="event-path">${safe(event.path)}</span>` : ''}
        </div>
      </article>
    </li>`).join('');

  const changes = state.events.filter((event) => event.kind === 'change');
  const agents = new Set(state.events.filter((event) => event.kind === 'agent').map((event) => event.actor));
  document.querySelector('#change-count').textContent = changes.length;
  document.querySelector('#agent-count').textContent = agents.size;
  document.querySelector('#risk-count').textContent = changes.filter((event) => event.risk === 'high' || event.risk === 'medium').length;
}

function selectEvent(sequence) {
  const event = state.events.find((candidate) => candidate.sequence === sequence);
  if (!event) return;
  state.selected = sequence;
  document.querySelector('#detail-empty').classList.add('hidden');
  document.querySelector('#detail-content').classList.remove('hidden');
  document.querySelector('#detail-confidence').textContent = event.confidence;
  document.querySelector('#detail-risk').textContent = `${event.risk} risk`;
  document.querySelector('#detail-title').textContent = event.title;
  document.querySelector('#detail-summary').textContent = event.summary;
  document.querySelector('#detail-path').textContent = event.path || 'No file path reported';
  document.querySelector('#detail-evidence').innerHTML = (event.evidence || []).map((item) => `<li>${safe(item)}</li>`).join('');
  render();
}

timeline.addEventListener('click', (event) => {
  const item = event.target.closest('.timeline-item');
  if (item) selectEvent(Number(item.dataset.sequence));
});
timeline.addEventListener('keydown', (event) => {
  const item = event.target.closest('.timeline-item');
  if (item && (event.key === 'Enter' || event.key === ' ')) {
    event.preventDefault();
    selectEvent(Number(item.dataset.sequence));
  }
});

document.querySelectorAll('.filter').forEach((button) => button.addEventListener('click', () => {
  state.filter = button.dataset.filter;
  document.querySelectorAll('.filter').forEach((candidate) => candidate.classList.toggle('active', candidate === button));
  render();
}));

async function start() {
  try {
    const response = await fetch('/api/state');
    if (!response.ok) throw new Error(`state request failed: ${response.status}`);
    const snapshot = await response.json();
    document.querySelector('#repo-name').textContent = snapshot.repository;
    state.events = snapshot.events || [];
    render();
  } catch (error) {
    connection.classList.add('offline');
    connection.querySelector('b').textContent = 'Unavailable';
    return;
  }

  const stream = new EventSource('/api/stream');
  stream.onopen = () => {
    connection.classList.remove('offline');
    connection.querySelector('b').textContent = 'Live';
  };
  stream.onmessage = (message) => {
    const event = JSON.parse(message.data);
    if (!state.events.some((candidate) => candidate.sequence === event.sequence)) {
      state.events.push(event);
      if (state.events.length > 500) state.events.shift();
      render();
    }
  };
  stream.onerror = () => {
    connection.classList.add('offline');
    connection.querySelector('b').textContent = 'Reconnecting';
  };
}

start();
setInterval(render, 10000);
