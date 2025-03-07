package sqlite

import (
	"context"
	"strings"

	"github.com/usememos/memos/store"
)

func (d *Driver) CreateUser(ctx context.Context, create *store.User) (*store.User, error) {
	stmt := `
		INSERT INTO user (
			username,
			role,
			email,
			nickname,
			password_hash
		)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, avatar_url, created_ts, updated_ts, row_status
	`
	if err := d.db.QueryRowContext(
		ctx,
		stmt,
		create.Username,
		create.Role,
		create.Email,
		create.Nickname,
		create.PasswordHash,
	).Scan(
		&create.ID,
		&create.AvatarURL,
		&create.CreatedTs,
		&create.UpdatedTs,
		&create.RowStatus,
	); err != nil {
		return nil, err
	}

	return create, nil
}

func (d *Driver) UpdateUser(ctx context.Context, update *store.UpdateUser) (*store.User, error) {
	set, args := []string{}, []any{}
	if v := update.UpdatedTs; v != nil {
		set, args = append(set, "updated_ts = ?"), append(args, *v)
	}
	if v := update.RowStatus; v != nil {
		set, args = append(set, "row_status = ?"), append(args, *v)
	}
	if v := update.Username; v != nil {
		set, args = append(set, "username = ?"), append(args, *v)
	}
	if v := update.Email; v != nil {
		set, args = append(set, "email = ?"), append(args, *v)
	}
	if v := update.Nickname; v != nil {
		set, args = append(set, "nickname = ?"), append(args, *v)
	}
	if v := update.AvatarURL; v != nil {
		set, args = append(set, "avatar_url = ?"), append(args, *v)
	}
	if v := update.PasswordHash; v != nil {
		set, args = append(set, "password_hash = ?"), append(args, *v)
	}
	args = append(args, update.ID)

	query := `
		UPDATE user
		SET ` + strings.Join(set, ", ") + `
		WHERE id = ?
		RETURNING id, username, role, email, nickname, password_hash, avatar_url, created_ts, updated_ts, row_status
	`
	user := &store.User{}
	if err := d.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Role,
		&user.Email,
		&user.Nickname,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.CreatedTs,
		&user.UpdatedTs,
		&user.RowStatus,
	); err != nil {
		return nil, err
	}

	return user, nil
}

func (d *Driver) ListUsers(ctx context.Context, find *store.FindUser) ([]*store.User, error) {
	where, args := []string{"1 = 1"}, []any{}

	if v := find.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := find.Username; v != nil {
		where, args = append(where, "username = ?"), append(args, *v)
	}
	if v := find.Role; v != nil {
		where, args = append(where, "role = ?"), append(args, *v)
	}
	if v := find.Email; v != nil {
		where, args = append(where, "email = ?"), append(args, *v)
	}
	if v := find.Nickname; v != nil {
		where, args = append(where, "nickname = ?"), append(args, *v)
	}

	query := `
		SELECT 
			id,
			username,
			role,
			email,
			nickname,
			password_hash,
			avatar_url,
			created_ts,
			updated_ts,
			row_status
		FROM user
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY created_ts DESC, row_status DESC
	`
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*store.User, 0)
	for rows.Next() {
		var user store.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Role,
			&user.Email,
			&user.Nickname,
			&user.PasswordHash,
			&user.AvatarURL,
			&user.CreatedTs,
			&user.UpdatedTs,
			&user.RowStatus,
		); err != nil {
			return nil, err
		}
		list = append(list, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (d *Driver) DeleteUser(ctx context.Context, delete *store.DeleteUser) error {
	result, err := d.db.ExecContext(ctx, `
		DELETE FROM user WHERE id = ?
	`, delete.ID)
	if err != nil {
		return err
	}
	if _, err := result.RowsAffected(); err != nil {
		return err
	}

	if err := d.Vacuum(ctx); err != nil {
		// Prevent linter warning.
		return err
	}

	return nil
}
