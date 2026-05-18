import React, { useEffect, useState } from 'react';
import Link from '@docusaurus/Link';

export default function GithubNavbarItem({ href, ...props }) {
  const [stats, setStats] = useState({ stars: 0, forks: 0, version: 'v0.1.1' });

  useEffect(() => {
    // Fetch repo data
    fetch('https://api.github.com/repos/balaji01-4d/pgxcli')
      .then(res => res.json())
      .then(data => {
        if (data.stargazers_count !== undefined) {
          setStats(prev => ({ ...prev, stars: data.stargazers_count, forks: data.forks_count }));
        }
      })
      .catch(console.error);

    fetch('https://api.github.com/repos/balaji01-4d/pgxcli/releases/latest')
      .then(res => res.json())
      .then(data => {
        if (data.tag_name) {
          setStats(prev => ({ ...prev, version: data.tag_name }));
        }
      })
      .catch(console.error);
  }, []);

  const formatNumber = (num) => {
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'k';
    }
    return num;
  };

  return (
    <Link to={href} className="navbar__item navbar__link custom-github-nav-item" {...props}>
      <div className="github-nav-container">
        <svg
            className="github-logo"
            viewBox="0 0 24 24"
            fill="currentColor"
            width="24"
            height="24"
          >
            <path d="M23.546 10.93 13.067.452c-.604-.603-1.582-.603-2.188 0L8.708 2.627l2.76 2.76c.645-.215 1.379-.07 1.889.441.516.515.658 1.258.438 1.9l2.775 2.776c.64-.22 1.383-.08 1.895.433.636.636.636 1.67 0 2.305-.636.636-1.671.636-2.305 0-.536-.536-.677-1.284-.446-1.928l-2.748-2.748v5.92c.26.173.493.407.671.688.544.85.306 1.983-.544 2.527-.85.544-1.983.306-2.527-.544-.544-.85-.306-1.983.544-2.527.18-.115.385-.19.596-.226v-5.918c-.22-.038-.435-.116-.626-.238L6.46 10.742c-.035.21-.113.415-.228.595-.544.85-1.677 1.088-2.527.544-.85-.544-1.088-1.677-.544-2.527.544-.85 1.677-1.088 2.527-.544.205.131.376.3.51.493l2.25-2.25L1.085 13.12c-.604.603-.604 1.582 0 2.188l10.48 10.479c.604.603 1.582.603 2.188 0l10.48-10.48c.603-.604.603-1.583-.001-2.188z" />
          </svg>
        <div className="github-nav-details">
          <div className="github-nav-title">GitHub</div>
          <div className="github-nav-stats">
            <span className="stat-item">
              <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor"><path fillRule="evenodd" d="M2.5 7.775V2.75a.25.25 0 01.25-.25h5.025a.25.25 0 01.177.073l6.25 6.25a.25.25 0 010 .354l-5.025 5.025a.25.25 0 01-.354 0l-6.25-6.25a.25.25 0 01-.073-.177zm-1.5 0V2.75C1 1.784 1.784 1 2.75 1h5.025c.464 0 .91.184 1.238.513l6.25 6.25a1.75 1.75 0 010 2.474l-5.026 5.026a1.75 1.75 0 01-2.474 0l-6.25-6.25A1.75 1.75 0 011 7.775zM6 5a1 1 0 100 2 1 1 0 000-2z"></path></svg>
              <span>{stats.version}</span>
            </span>
            <span className="stat-item">
              <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor"><path fillRule="evenodd" d="M8 .25a.75.75 0 01.673.418l1.882 3.815 4.21.312a.75.75 0 01.416 1.279l-3.046 2.97.719 4.192a.75.75 0 01-1.088.791L8 12.347l-3.766 1.98a.75.75 0 01-1.088-.79l.72-4.194L.818 6.074a.75.75 0 01.416-1.28l4.21-.311L7.327.668A.75.75 0 018 .25zm0 2.445L6.615 5.5a.75.75 0 01-.564.41l-3.097.23 2.24 2.184a.75.75 0 01.216.664l-.528 3.084 2.769-1.456a.75.75 0 01.698 0l2.77 1.456-.53-3.084a.75.75 0 01.216-.664l2.24-2.183-3.096-.23a.75.75 0 01-.564-.41L8 2.694v.001z"></path></svg>
              <span>{formatNumber(stats.stars)}</span>
            </span>
            <span className="stat-item">
              <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor"><path fillRule="evenodd" d="M5 3.25a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm0 2.122a2.25 2.25 0 10-1.5 0v.878A2.25 2.25 0 005.75 8.5h1.5v2.128a2.251 2.251 0 101.5 0V8.5h1.5a2.25 2.25 0 002.25-2.25v-.878a2.25 2.25 0 10-1.5 0v.878a.75.75 0 01-.75.75h-4.5A.75.75 0 015 6.25v-.878zm3.75 7.378a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm3-8.75a.75.75 0 100-1.5.75.75 0 000 1.5z"></path></svg>
              <span>{formatNumber(stats.forks)}</span>
            </span>
          </div>
        </div>
      </div>
    </Link>
  );
}
