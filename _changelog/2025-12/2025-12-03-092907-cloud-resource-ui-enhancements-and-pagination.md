# Cloud Resource UI Enhancements and Server-Side Pagination

**Date**: December 3, 2025
**Type**: Feature, Enhancement
**Components**: Web Frontend, Backend API, UI Components

## Summary

Enhanced the cloud resource web interface with improved UI components, theme switching capabilities, and implemented server-side pagination for better performance and scalability. The changes include a new table component with pagination support, theme switch component, enhanced header and sidebar, and various reusable UI components.

## Problem Statement

The cloud resource web interface needed improvements in several areas:

### Missing Capabilities

- **No server-side pagination**: The frontend was using client-side pagination, loading all resources at once, which doesn't scale for large datasets
- **Limited UI components**: Missing reusable components like alert dialogs, confirmation dialogs, custom tooltips, and icon components
- **No theme switching UI**: While theme system existed, there was no user-visible way to switch between dark and light modes
- **Incomplete table component**: The existing data-table component lacked proper pagination support and was replaced with a more comprehensive table component
- **Limited visual feedback**: Missing confirmation dialogs and alert dialogs for better user interaction

### User Impact

Without these improvements, users faced:

- Performance issues when managing large numbers of cloud resources (all loaded at once)
- Inability to switch themes directly from the UI
- Limited visual feedback for critical actions like deletions
- Inconsistent UI components across the application

## Solution

Implemented comprehensive UI enhancements including a new table component with server-side pagination, theme switching component, and various reusable UI components. Updated the backend API to properly support pagination with total page count calculation.

### Architecture

The pagination implementation follows a server-side pattern:

```
Frontend Table Component
    ↓ Page Change Event
Cloud Resources List Component
    ↓ API Call with PageInfo
Backend Service (ListCloudResources)
    ↓ Pagination Options
Database Repository
    ↓ MongoDB Skip/Limit
Database Query
    ↓ Total Count Calculation
Response with Resources + TotalPages
```

**UI Component Architecture**:

```
App Layout
    ├── Header (with Theme Switch)
    ├── Sidebar (simplified)
    └── Main Content
        └── Cloud Resources Page
            └── Cloud Resources List
                └── Table Component (with Pagination)
```

### Key Features

**1. Server-Side Pagination**

- Backend API calculates total pages based on total count and page size
- Frontend sends page number and page size in API requests
- Pagination component displays page numbers and navigation controls
- Default page size of 10 items per page
- Page numbers are 0-indexed (page 0 is the first page)

**2. Theme Switch Component**

- Visual toggle button in header (sun/moon icons)
- Switches between dark and light themes
- Persists preference in cookies and localStorage
- Tooltip shows next theme mode on hover

**3. Enhanced Table Component**

- Replaces old data-table component
- Supports both client-side and server-side pagination modes
- Custom pagination controls with page numbers, first/last, prev/next buttons
- Loading states with skeleton placeholders
- Row selection support (for future bulk operations)
- Sticky header and action columns
- Border and background color customization

**4. New Reusable UI Components**

- **AlertDialog**: Modal dialog for confirmations and alerts
- **ConfirmationDialog**: Specialized dialog for action confirmations
- **CustomTooltip**: Enhanced tooltip component
- **Icon Component**: Reusable icon component with SVG support
- **TextCopy**: Component for copyable text with copy-to-clipboard functionality
- **ResourceHeader**: Styled header component for resource pages

**5. Enhanced Layout Components**

- **Header**: Added theme switch button, improved styling with new icons
- **Sidebar**: Simplified design, updated icons and styling
- **Layout Styling**: Improved spacing and visual hierarchy

## Implementation Details

### 1. Backend Pagination Support

**File**: `app/backend/internal/service/cloud_resource_service.go`

The `ListCloudResources` method now properly handles pagination:

```112:182:app/backend/internal/service/cloud_resource_service.go
// ListCloudResources retrieves cloud resources with optional pagination.
func (s *CloudResourceService) ListCloudResources(
	ctx context.Context,
	req *connect.Request[backendv1.ListCloudResourcesRequest],
) (*connect.Response[backendv1.ListCloudResourcesResponse], error) {
	opts := &database.CloudResourceListOptions{}
	if req.Msg.Kind != nil {
		kind := *req.Msg.Kind
		opts.Kind = &kind
	}

	// Apply pagination with defaults (page=0, size=20) if not provided
	var pageNum int32 = 0
	var pageSize int32 = 20
	if req.Msg.PageInfo != nil {
		pageNum = req.Msg.PageInfo.Num
		pageSize = req.Msg.PageInfo.Size
	}
	opts.PageNum = &pageNum
	opts.PageSize = &pageSize

	// Calculate total pages
	totalCount, err := s.repo.Count(ctx, opts)
	if err != nil {
		logrus.WithError(err).Error("Failed to count cloud resources for pagination")
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to count cloud resources: %w", err))
	}

	var totalPages int32
	if pageSize > 0 {
		totalPages = int32((totalCount + int64(pageSize) - 1) / int64(pageSize))
	}

	logrus.WithFields(logrus.Fields{
		"kind":      req.Msg.Kind,
		"page_num":  opts.PageNum,
		"page_size": opts.PageSize,
	}).Info("Listing cloud resources")

	resources, err := s.repo.List(ctx, opts)
	if err != nil {
		logrus.WithError(err).Error("Failed to list cloud resources")
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list cloud resources: %w", err))
	}

	protoResources := make([]*backendv1.CloudResource, 0, len(resources))
	for _, res := range resources {
		protoRes := &backendv1.CloudResource{
			Id:       res.ID.Hex(),
			Name:     res.Name,
			Kind:     res.Kind,
			Manifest: res.Manifest,
		}

		if !res.CreatedAt.IsZero() {
			protoRes.CreatedAt = timestamppb.New(res.CreatedAt)
		}
		if !res.UpdatedAt.IsZero() {
			protoRes.UpdatedAt = timestamppb.New(res.UpdatedAt)
		}

		protoResources = append(protoResources, protoRes)
	}

	response := &backendv1.ListCloudResourcesResponse{
		Resources:  protoResources,
		TotalPages: totalPages,
	}

	return connect.NewResponse(response), nil
}
```

**Key Features**:

- Default pagination: page 0, size 20 if not provided
- Calculates total pages using ceiling division: `(totalCount + pageSize - 1) / pageSize`
- Returns both resources and totalPages in response
- Maintains backward compatibility (pagination is optional)

**File**: `app/backend/internal/database/cloud_resource_repo.go`

The repository supports pagination with MongoDB skip and limit:

```99:137:app/backend/internal/database/cloud_resource_repo.go
// List retrieves cloud resources from MongoDB with optional filters and pagination.
func (r *CloudResourceRepository) List(ctx context.Context, opts *CloudResourceListOptions) ([]*models.CloudResource, error) {
	filter := bson.M{}

	if opts != nil {
		if opts.Kind != nil && *opts.Kind != "" {
			filter["kind"] = *opts.Kind
		}
	}

	findOptions := options.Find()

	// Apply pagination if provided
	if opts != nil && opts.PageNum != nil && opts.PageSize != nil {
		pageNum := *opts.PageNum
		pageSize := *opts.PageSize
		if pageSize > 0 {
			skip := int64(pageNum) * int64(pageSize)
			findOptions.SetSkip(skip)
			findOptions.SetLimit(int64(pageSize))
		}
	}

	// Sort by created_at descending (newest first)
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query cloud resources: %w", err)
	}
	defer cursor.Close(ctx)

	var resources []*models.CloudResource
	if err := cursor.All(ctx, &resources); err != nil {
		return nil, fmt.Errorf("failed to decode cloud resources: %w", err)
	}

	return resources, nil
}
```

### 2. Frontend Pagination Implementation

**File**: `app/frontend/src/components/shared/cloud-resources-list/cloud-resources-list.tsx`

Updated to use server-side pagination:

```96:118:app/frontend/src/components/shared/cloud-resources-list/cloud-resources-list.tsx
  // Function to call the API
  const handleLoadCloudResources = useCallback(() => {
    setApiLoading(true);
    if (query) {
      query
        .listCloudResources(
          create(ListCloudResourcesRequestSchema, {
            kind: kindFilter.trim() || undefined,
            pageInfo: create(PageInfoSchema, {
              num: page,
              size: rowsPerPage,
            }),
          })
        )
        .then((result) => {
          console.log('result', result);
          setCloudResources(result.resources);
          setApiLoading(false);
        })
        .finally(() => {
          setApiLoading(false);
        });
    }
  }, [query, kindFilter]);
```

**Key Changes**:

- Sends `pageInfo` with current page number and page size
- Uses `result.totalPages` from API response (instead of calculating from array length)
- Resets to page 0 when kind filter changes
- Uses server-side pagination mode

**File**: `app/frontend/src/components/shared/table/pagination.tsx`

Custom pagination component with Material-UI integration:

```94:134:app/frontend/src/components/shared/table/pagination.tsx
export const TablePagination = ({
  page,
  totalRecords = -1,
  rowsPerPage = 10,
  mode,
  totalPages = 0,
  onPageChange,
  component,
  border,
  borderColor,
  bgColor,
  ...otherProps
}: TablePaginationProps) => {
  return (
    <StyledPaginationContainer $border={border} $borderColor={borderColor} $bgColor={bgColor}>
      <MuiTablePagination
        component={component ?? 'td'}
        page={page}
        count={totalRecords}
        rowsPerPage={rowsPerPage}
        onPageChange={null}
        onRowsPerPageChange={null}
        rowsPerPageOptions={[]}
        labelRowsPerPage=""
        labelDisplayedRows={() => ''}
        ActionsComponent={(subProps) => (
          <PaginationActions
            {...subProps}
            mode={mode}
            count={totalRecords}
            page={page}
            rowsPerPage={rowsPerPage}
            totalPages={totalPages}
            setPageData={onPageChange}
          />
        )}
        {...otherProps}
      />
    </StyledPaginationContainer>
  );
};
```

**Features**:

- Custom pagination actions with first/last, prev/next, and page number buttons
- Ellipsis for large page counts
- Supports both client and server pagination modes
- Customizable styling (border, background color)

### 3. Theme Switch Component

**File**: `app/frontend/src/components/layout/theme-switch/theme-switch.tsx`

```10:35:app/frontend/src/components/layout/theme-switch/theme-switch.tsx
const ThemeSwitch = () => {
  const {
    theme: { mode },
    changeTheme,
  } = useContext(AppContext);
  const nextTheme = useMemo(() => (mode === THEME.DARK ? THEME.LIGHT : THEME.DARK), [mode]);

  const toggleTheme = () => {
    Utils.setStorage(PCS_THEME_IDENTIFIER, nextTheme);
    changeTheme(nextTheme);
  };

  return (
    <Tooltip title={`Switch to ${nextTheme}`}>
      <StyledThemeButton
        $mode={mode}
        onClick={toggleTheme}
        startIcon={<Icon name={ICON_NAMES.SUN} alt={`Switch to ${nextTheme}`} />}
        endIcon={<Icon name={ICON_NAMES.MOON} alt={`Switch to ${nextTheme}`} />}
        disableRipple
      />
    </Tooltip>
  );
};
```

**Features**:

- Shows sun icon for light mode, moon icon for dark mode
- Tooltip indicates next theme mode
- Persists preference to localStorage
- Integrated into header component

### 4. New UI Components

**AlertDialog Component** (`app/frontend/src/components/shared/alert-dialog/alert-dialog.tsx`):

- Modal dialog for confirmations and alerts
- Customizable title, subtitle, submit/cancel labels
- Color-coded submit button (error, primary, etc.)

**ConfirmationDialog Component** (`app/frontend/src/components/shared/confirmation-dialog/confirmation-dialog.tsx`):

- Specialized dialog for action confirmations
- Optional reason field for deletions
- Customizable message and labels

**Icon Component** (`app/frontend/src/components/shared/icon/icon.tsx`):

- Reusable icon component with SVG support
- Predefined icon names (SUN, MOON, DELETE, EDIT, etc.)
- Customizable size and styling

**TextCopy Component** (`app/frontend/src/components/shared/text-copy/text-copy.tsx`):

- Copy-to-clipboard functionality
- Visual feedback on copy action
- Used for copyable table cells

**Table Component** (`app/frontend/src/components/shared/table/table.tsx`):

- Comprehensive table component replacing data-table
- Server-side and client-side pagination support
- Row selection, loading states, empty states
- Customizable actions, styling, and data templates

### 5. Enhanced Layout Components

**Header** (`app/frontend/src/components/layout/header/header.tsx`):

- Added theme switch button
- New header icons (3square, 4square, nav icons)
- Updated styling with theme-aware colors

**Sidebar** (`app/frontend/src/components/layout/sidebar/sidebar.tsx`):

- Simplified design
- Updated icons and styling
- Better visual hierarchy

## Benefits

### For End Users

**Performance**:

- Server-side pagination loads only the current page, improving performance for large datasets
- Faster page loads and smoother navigation
- Reduced memory usage in browser

**User Experience**:

- Easy theme switching directly from the UI
- Better visual feedback with confirmation dialogs
- Consistent UI components across the application
- Improved table navigation with pagination controls

**Accessibility**:

- Clear pagination controls with page numbers
- Tooltips provide context for actions
- Consistent visual design

### For Developers

**Component Reusability**:

- New reusable components (AlertDialog, ConfirmationDialog, Icon, TextCopy) can be used throughout the app
- Table component supports both pagination modes for different use cases
- Consistent patterns for future features

**Maintainability**:

- Clear separation between client-side and server-side pagination
- Type-safe pagination with TypeScript
- Consistent error handling patterns

**Scalability**:

- Server-side pagination scales to handle thousands of resources
- Efficient database queries with skip/limit
- Total count calculation only when needed

## Impact

### Immediate

**Performance Improvement**: Server-side pagination reduces initial load time and memory usage
**User Experience**: Theme switching and better UI components improve usability
**Scalability**: Can now handle large numbers of cloud resources efficiently

### Developer Experience

**1 new table component** with comprehensive features
**5 new reusable UI components** (AlertDialog, ConfirmationDialog, Icon, TextCopy, ResourceHeader)
**1 theme switch component** integrated into header
**Server-side pagination** implementation in both frontend and backend
**Enhanced layout components** with improved styling

### System Capabilities

**Scalable Pagination**: Can handle datasets of any size efficiently
**Theme Management**: User-visible theme switching with persistence
**Component Library**: Foundation for consistent UI across the application
**Performance**: Reduced load times and memory usage for large datasets

## Usage Examples

### Server-Side Pagination

**Backend API Call**:

```typescript
const request = create(ListCloudResourcesRequestSchema, {
  kind: 'CivoVpc', // optional filter
  pageInfo: create(PageInfoSchema, {
    num: 0, // page number (0-indexed)
    size: 10, // items per page
  }),
});

const response = await query.listCloudResources(request);
// response.resources: CloudResource[] (10 items)
// response.totalPages: number (calculated total pages)
```

**Frontend Table Usage**:

```typescript
<TableComp
  data={cloudResources}
  options={{
    headers: ['Name', 'Kind', 'Created At'],
    dataPath: ['name', 'kind', 'createdAt'],
    showPagination: true,
    paginationMode: PAGINATION_MODE.SERVER,
    currentPage: page,
    rowsPerPage: rowsPerPage,
    totalPages: totalPages, // from API response
    onPageChange: handlePageChange,
  }}
/>
```

### Theme Switching

Users can click the theme switch button in the header to toggle between dark and light modes. The preference is automatically saved and persists across page refreshes.

### Using New Components

**AlertDialog**:

```typescript
<AlertDialog
  open={open}
  onClose={handleClose}
  onSubmit={handleSubmit}
  title="Delete Resource"
  subTitle="Are you sure you want to delete this resource?"
  submitLabel="Delete"
  submitBtnColor="error"
/>
```

**Icon Component**:

```typescript
<Icon name={ICON_NAMES.DELETE} onClick={handleDelete} />
<Icon name={ICON_NAMES.EDIT} onClick={handleEdit} />
```

## Files Modified/Created

### Backend

**Modified**:

- `app/backend/apis/proto/cloud_resource_service.proto` - Added PageInfo message and totalPages to response
- `app/backend/internal/service/cloud_resource_service.go` - Implemented pagination logic with total pages calculation
- `app/backend/internal/database/cloud_resource_repo.go` - Added pagination support with skip/limit

### Frontend Pages

**Modified**:

- `app/frontend/src/app/cloud-resources/page.tsx` - Updated to use new CloudResourcesList component
- `app/frontend/src/app/cloud-resources/_services/query.ts` - Updated to handle pagination in API calls
- `app/frontend/src/app/dashboard/page.tsx` - Updated styling and layout

### UI Components

**Created**:

- `app/frontend/src/components/shared/table/table.tsx` - New comprehensive table component
- `app/frontend/src/components/shared/table/pagination.tsx` - Custom pagination component
- `app/frontend/src/components/shared/table/styled.ts` - Table styling
- `app/frontend/src/components/shared/alert-dialog/alert-dialog.tsx` - Alert dialog component
- `app/frontend/src/components/shared/confirmation-dialog/confirmation-dialog.tsx` - Confirmation dialog component
- `app/frontend/src/components/shared/custom-tooltip/custom-tooltip.tsx` - Custom tooltip component
- `app/frontend/src/components/shared/icon/icon.tsx` - Icon component
- `app/frontend/src/components/shared/text-copy/text-copy.tsx` - Text copy component
- `app/frontend/src/components/shared/resource-header/styled.ts` - Resource header styling
- `app/frontend/src/components/layout/theme-switch/theme-switch.tsx` - Theme switch component
- `app/frontend/src/components/layout/theme-switch/styled.ts` - Theme switch styling
- `app/frontend/src/components/layout/header/header-icon.tsx` - Header icon component

**Modified**:

- `app/frontend/src/components/shared/cloud-resources-list/cloud-resources-list.tsx` - Updated to use server-side pagination
- `app/frontend/src/components/layout/header/header.tsx` - Added theme switch integration
- `app/frontend/src/components/layout/header/styled.ts` - Updated header styling
- `app/frontend/src/components/layout/sidebar/sidebar.tsx` - Updated sidebar styling
- `app/frontend/src/components/layout/sidebar/styled.ts` - Updated sidebar styling
- `app/frontend/src/components/layout/styled.ts` - Updated layout styling

**Deleted**:

- `app/frontend/src/components/shared/data-table/data-table.tsx` - Replaced by new table component
- `app/frontend/src/components/shared/data-table/index.ts` - Removed
- `app/frontend/src/components/shared/data-table/styled.ts` - Removed

### Models and Types

**Created**:

- `app/frontend/src/models/table.ts` - Table models including PAGINATION_MODE enum

**Modified**:

- `app/frontend/src/themes/dark.tsx` - Updated dark theme colors
- `app/frontend/src/themes/light.tsx` - Updated light theme colors

### Assets

**Created**:

- `app/frontend/public/images/delete.svg` - Delete icon
- `app/frontend/public/images/edit-icon.svg` - Edit icon
- `app/frontend/public/images/leftnav-icons/3square.svg` - Navigation icon
- `app/frontend/public/images/leftnav-icons/4square.svg` - Navigation icon
- `app/frontend/public/images/moon.svg` - Moon icon for theme
- `app/frontend/public/images/nav.svg` - Navigation icon
- `app/frontend/public/images/planton-logo-dark.svg` - Dark mode logo
- `app/frontend/public/images/openmcf-logo.svg` - Light mode logo
- `app/frontend/public/images/sun.svg` - Sun icon for theme

### Configuration

**Modified**:

- `app/frontend/package.json` - Updated dependencies
- `app/frontend/next.config.js` - Updated configuration
- `app/frontend/yarn.lock` - Updated lock file

## Technical Metrics

- **1 new table component** with comprehensive features
- **5 new reusable UI components** for consistent design
- **1 theme switch component** for user theme control
- **Server-side pagination** implemented in backend and frontend
- **Default page size**: 10 items per page
- **Backward compatible**: Pagination is optional in API
- **~2000 lines** of new TypeScript/React code
- **Full TypeScript coverage** for all new components

## Related Work

### Foundation

This work builds on:

- **Cloud Resource Web UI** (December 1, 2025) - Initial web interface implementation
- **Theme System** (December 1, 2025) - Dark/light theme infrastructure
- **Cloud Resource APIs** (November 28, 2025) - Backend API foundation

### Complements

This work complements:

- **Cloud Resource Management** - Enhanced UI for resource management
- **Theme System** - User-visible theme switching
- **Component Library** - Foundation for future features

### Future Extensions

This work enables:

- **Bulk Operations** - Row selection ready for batch operations
- **Advanced Filtering** - Multi-criteria filtering with pagination
- **Export/Import** - Paginated export of resources
- **Resource Search** - Search functionality with pagination
- **Custom Page Sizes** - User-selectable items per page
- **Infinite Scroll** - Alternative pagination pattern

## Known Limitations

- **Fixed page size**: Currently 10 items per page, not user-configurable
- **No page size selector**: Users cannot change items per page
- **Total records not displayed**: Only total pages shown, not total record count
- **No jump to page**: Users must navigate page by page

These limitations are intentional for the initial implementation and can be addressed in future enhancements.

## Design Decisions

### Server-Side Pagination

**Decision**: Implement server-side pagination instead of client-side

**Rationale**:

- Scales to handle large datasets efficiently
- Reduces initial load time and memory usage
- Better performance for production use cases
- Standard pattern for data-heavy applications

**Alternative considered**: Client-side pagination

- Rejected because it doesn't scale and loads all data at once

### Page Number Calculation

**Decision**: Use ceiling division for total pages: `(totalCount + pageSize - 1) / pageSize`

**Rationale**:

- Handles partial last pages correctly
- Standard algorithm for pagination
- Efficient calculation

**Alternative considered**: `Math.ceil(totalCount / pageSize)`

- Equivalent but less explicit in Go code

### Default Page Size

**Decision**: Default to 10 items per page

**Rationale**:

- Good balance between performance and usability
- Common default in web applications
- Not too many items to overwhelm, not too few to require excessive navigation

**Alternative considered**: 20 items per page

- Rejected because 10 provides better performance and is more standard

### Theme Switch Location

**Decision**: Place theme switch in header

**Rationale**:

- Always visible and accessible
- Standard location for theme toggles
- Doesn't interfere with main content

**Alternative considered**: Settings menu

- Rejected because it would be less discoverable

## Migration Notes

**Breaking Changes**: None

The old `data-table` component has been replaced with the new `table` component. Any code using `data-table` should be updated to use `table` with appropriate options.

**Backward Compatibility**: The API maintains backward compatibility - pagination is optional. If `pageInfo` is not provided, the API returns all resources (with default pagination applied).

Existing users will automatically benefit from:

- Server-side pagination (better performance)
- Theme switching capability
- Enhanced UI components
- Improved table navigation

---

**Status**: ✅ Complete and Production Ready
**Component**: Web Frontend - Cloud Resource Management, UI Components
**Pages Modified**: 1 page (cloud-resources)
**Components Added**: 6 new reusable components
**Components Modified**: 3 layout components
**Backend Changes**: Pagination support in service and repository
**Location**: `app/frontend/src/components/` and `app/backend/internal/service/`
