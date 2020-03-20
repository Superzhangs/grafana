import React, { FC } from 'react';
import { css, cx } from 'emotion';
import { Icon } from '@grafana/ui';
import { IconType } from '@grafana/ui/src/components/Icon/types';
import { DashboardSection, DashboardSectionItem } from '../types';
import { SearchResultsItem } from './SearchResultsItem';
import { SearchResultsCheckbox } from './SearchResultsCheckbox';

type clickWithEvent = (item: DashboardSectionItem | DashboardSection, event: any) => void;
interface Props {
  results: DashboardSection[] | undefined;
  onSelectionChanged: any;
  onTagSelected: any;
  onFolderExpanding: any;
  onToggleSelection: clickWithEvent;
  editable: boolean;
}

export const SearchResults: FC<Props> = ({
  results,
  onSelectionChanged,
  onTagSelected,
  onFolderExpanding,
  onToggleSelection,
  editable,
}) => {
  const toggleFolderExpand = (section: DashboardSection) => {
    if (section.toggle) {
      if (!section.expanded && typeof onFolderExpanding === 'function') {
        onFolderExpanding();
      }

      section.toggle(section).then(() => {
        if (typeof onSelectionChanged === 'function') {
          onSelectionChanged();
        }
      });
    }
  };

  return !results || !results.length ? (
    <div className="search-results">
      <em className="muted">No dashboards found.</em>
    </div>
  ) : (
    <>
      {results.map(section => (
        <div className="search-section" key={section.id}>
          {!section.hideHeader ? (
            <div
              className={cx('search-section__header pointer', { selected: section.checked })}
              onClick={() => toggleFolderExpand(section)}
            >
              <SearchResultsCheckbox
                editable={editable}
                checked={section.checked}
                onClick={(e: MouseEvent) => onToggleSelection(section, e)}
              />
              <Icon
                className={css`
                  padding: 5px 0;
                  width: 43px;
                `}
                name={section.icon as IconType}
              />

              <span className="search-section__header__text">{section.title}</span>
              {section.url && (
                <a href={section.url} className="search-section__header__link">
                  <Icon name="cog" />
                </a>
              )}
              <Icon name={section.expanded ? 'angle-down' : 'angle-right'} className="search-section__header__toggle" />
            </div>
          ) : (
            <div className="search-section__header" />
          )}
          {section.expanded &&
            section.items.map(item => (
              <SearchResultsItem key={item.id} {...{ item, editable, onToggleSelection, onTagSelected }} />
            ))}
        </div>
      ))}
    </>
  );
};
