mod i18n;
mod intents;
mod models;
mod truco_core;
mod window;

use std::cell::RefCell;
use std::rc::Rc;

use adw::prelude::*;
use gtk::gdk;
use gtk::glib;

use i18n::{text, Locale};
use intents::{
    to_json, AppIntent, CreateHostPayload, GameActionPayload, HostVotePayload, JoinSessionPayload,
    NewOfflineGamePayload, ReplacementInvitePayload, SendChatPayload, SetLocalePayload,
};
use models::{AppEvent, GameSnapshot, SnapshotBundle};
use truco_core::TrucoCore;

struct AppState {
    pub core: Option<TrucoCore>,
    pub window: window::TrucoWindow,
    pub last_snapshot_str: String,
    pub locale: Locale,
}

fn main() {
    let app = adw::Application::builder()
        .application_id("dev.truco.Native")
        .build();

    app.connect_startup(|_| load_css());

    app.connect_activate(|app| {
        let window = window::TrucoWindow::new(app);
        let locale = Locale::PtBr;
        let core = TrucoCore::new();
        let state = Rc::new(RefCell::new(AppState {
            core: core.clone().ok(),
            window: window.clone(),
            last_snapshot_str: String::new(),
            locale,
        }));

        apply_locale(&window, locale);
        connect_shell_actions(&state);
        connect_primary_actions(&state);
        bind_poll_loop(&state);

        match core {
            Ok(core) => {
                set_status(
                    &window,
                    &format!(
                        "{}: {}",
                        text(locale, "table-ready"),
                        core.library_path().display()
                    ),
                );
                hide_banner(&window);
            }
            Err(err) => {
                let message = if err.to_string().contains("not found") {
                    text(locale, "missing-core").to_string()
                } else {
                    format!("{} {}", text(locale, "bad-core"), err)
                };
                show_banner(&window, &message);
                set_status(&window, &message);
                disable_session_actions(&window);
            }
        }

        adw::prelude::GtkWindowExt::present(&window);
    });

    app.run();
}

fn connect_shell_actions(state: &Rc<RefCell<AppState>>) {
    let window = state.borrow().window.clone();
    window.btn_banner_close().connect_clicked(glib::clone!(
        #[weak]
        window,
        move |_| {
            hide_banner(&window);
        }
    ));

    let copy_state = state.clone();
    window.btn_copy_invite().connect_clicked(move |_| {
        let state_ref = copy_state.borrow();
        if let Some(display) = gdk::Display::default() {
            let clipboard = display.clipboard();
            let key = state_ref
                .window
                .lbl_invite_key_display()
                .label()
                .replace("Chave: ", "")
                .replace("Key: ", "");
            if !key.trim().is_empty() && !key.contains("(Convidado)") {
                clipboard.set_text(&key);
                push_toast(&state_ref.window, text(state_ref.locale, "copied"));
            }
        }
    });

    let locale_state = state.clone();
    window.dd_locale().connect_selected_notify(move |dd| {
        let mut state_ref = locale_state.borrow_mut();
        state_ref.locale = if dd.selected() == 1 {
            Locale::EnUs
        } else {
            Locale::PtBr
        };
        apply_locale(&state_ref.window, state_ref.locale);
        let intent = AppIntent::with_payload(
            "set_locale",
            SetLocalePayload {
                locale: state_ref.locale.code(),
            },
        );
        dispatch_serialized(&mut state_ref, &intent, None);
    });
}

fn connect_primary_actions(state: &Rc<RefCell<AppState>>) {
    let window = state.borrow().window.clone();

    let offline_state = state.clone();
    window.btn_start_demo().connect_clicked(move |_| {
        let mut state_ref = offline_state.borrow_mut();
        let player_name = fallback_name(
            state_ref.window.entry_player_name().text().as_str(),
            state_ref.locale,
        );
        let num_players = if state_ref.window.dd_num_players().selected() == 1 {
            4
        } else {
            2
        };
        let (player_names, cpu_flags) = if num_players == 4 {
            (
                vec![
                    player_name,
                    "CPU-Direita".to_string(),
                    "CPU-Parceiro".to_string(),
                    "CPU-Esquerda".to_string(),
                ],
                vec![false, true, true, true],
            )
        } else {
            (
                vec![player_name, "CPU-Oponente".to_string()],
                vec![false, true],
            )
        };
        let locale_intent = AppIntent::with_payload(
            "set_locale",
            SetLocalePayload {
                locale: state_ref.locale.code(),
            },
        );
        dispatch_serialized(&mut state_ref, &locale_intent, None);
        let intent = AppIntent::with_payload(
            "new_offline_game",
            NewOfflineGamePayload {
                player_names,
                cpu_flags,
            },
        );
        let fallback_error = text(state_ref.locale, "connection-error").to_string();
        dispatch_serialized(&mut state_ref, &intent, Some(fallback_error.as_str()));
    });

    let host_state = state.clone();
    window.btn_host_online().connect_clicked(move |_| {
        let mut state_ref = host_state.borrow_mut();
        let host_name = fallback_name(
            state_ref.window.entry_player_name().text().as_str(),
            state_ref.locale,
        );
        let num_players = if state_ref.window.dd_num_players().selected() == 1 {
            4
        } else {
            2
        };
        let relay_url_value = state_ref.window.entry_relay_url().text().to_string();
        let relay_url = if relay_url_value.trim().is_empty() {
            None
        } else {
            Some(relay_url_value.trim())
        };
        let intent = AppIntent::with_payload(
            "create_host_session",
            CreateHostPayload {
                host_name: &host_name,
                num_players,
                relay_url,
            },
        );
        let fallback_error = text(state_ref.locale, "connection-error").to_string();
        dispatch_serialized(&mut state_ref, &intent, Some(fallback_error.as_str()));
    });

    let join_state = state.clone();
    window.btn_join_online().connect_clicked(move |_| {
        let mut state_ref = join_state.borrow_mut();
        let name = fallback_name(
            state_ref.window.entry_player_name().text().as_str(),
            state_ref.locale,
        );
        let key = state_ref.window.entry_invite_key().text().to_string();
        if key.trim().is_empty() {
            show_banner(&state_ref.window, "Informe a chave de convite.");
            return;
        }
        let desired_role = match state_ref.window.dd_desired_role().selected() {
            1 => Some("partner"),
            2 => Some("opponent"),
            _ => None,
        };
        let intent = AppIntent::with_payload(
            "join_session",
            JoinSessionPayload {
                key: key.trim(),
                player_name: &name,
                desired_role,
            },
        );
        let fallback_error = text(state_ref.locale, "connection-error").to_string();
        dispatch_serialized(&mut state_ref, &intent, Some(fallback_error.as_str()));
    });

    let start_online_state = state.clone();
    window.btn_start_online_match().connect_clicked(move |_| {
        let mut state_ref = start_online_state.borrow_mut();
        let intent = AppIntent::without_payload("start_hosted_match");
        let fallback_error = text(state_ref.locale, "connection-error").to_string();
        dispatch_serialized(&mut state_ref, &intent, Some(fallback_error.as_str()));
    });

    let leave_state = state.clone();
    window.btn_leave_online().connect_clicked(move |_| {
        let mut state_ref = leave_state.borrow_mut();
        let intent = AppIntent::without_payload("close_session");
        dispatch_serialized(&mut state_ref, &intent, None);
        clear_listbox(&state_ref.window.list_chat());
    });

    let leave_match_state = state.clone();
    window.btn_leave_match().connect_clicked(move |_| {
        let mut state_ref = leave_match_state.borrow_mut();
        let intent = AppIntent::without_payload("close_session");
        dispatch_serialized(&mut state_ref, &intent, None);
        state_ref.window.game_over_overlay().set_visible(false);
        state_ref.window.trick_end_overlay().set_visible(false);
        state_ref
            .window
            .main_stack()
            .set_visible_child_name("lobby");
        clear_listbox(&state_ref.window.list_chat());
    });

    let chat_state = state.clone();
    window
        .btn_send_chat()
        .connect_clicked(move |_| send_chat(&chat_state));
    let chat_state_enter = state.clone();
    window
        .entry_chat()
        .connect_activate(move |_| send_chat(&chat_state_enter));

    let back_state = state.clone();
    window.btn_back_lobby().connect_clicked(move |_| {
        let state_ref = back_state.borrow();
        state_ref.window.game_over_overlay().set_visible(false);
        state_ref
            .window
            .main_stack()
            .set_visible_child_name("lobby");
    });
}

fn bind_poll_loop(state: &Rc<RefCell<AppState>>) {
    let state = state.clone();
    glib::timeout_add_local(std::time::Duration::from_millis(50), move || {
        let mut app = state.borrow_mut();
        let Some(core) = app.core.clone() else {
            return glib::ControlFlow::Continue;
        };

        match core.poll_event() {
            Ok(Some(event_str)) => {
                if let Ok(ev) = serde_json::from_str::<AppEvent>(&event_str) {
                    process_event(&mut app, &ev);
                }
            }
            Ok(None) => {}
            Err(err) => {
                show_banner(&app.window, &err.to_string());
            }
        }

        match core.snapshot() {
            Ok(Some(snapshot_str)) => {
                if snapshot_str != app.last_snapshot_str {
                    app.last_snapshot_str = snapshot_str.clone();
                    if let Some(bundle) = SnapshotBundle::from_json(&snapshot_str) {
                        if let Some(locale_code) = bundle.locale.as_deref() {
                            app.locale = Locale::from_code(locale_code);
                            apply_locale(&app.window, app.locale);
                        }
                        if let Some(conn) = bundle.connection.as_ref() {
                            if let Some(err) =
                                conn.last_error.as_ref().and_then(|e| e.message.as_deref())
                            {
                                show_banner(&app.window, err);
                            }
                        }
                        update_ui(&app.window, &bundle, &core, app.locale);
                    }
                }
            }
            Ok(None) => {}
            Err(err) => show_banner(&app.window, &err.to_string()),
        }

        glib::ControlFlow::Continue
    });
}

fn send_chat(state: &Rc<RefCell<AppState>>) {
    let mut app = state.borrow_mut();
    let text_value = app.window.entry_chat().text().to_string();
    if text_value.trim().is_empty() {
        return;
    }
    let intent = AppIntent::with_payload(
        "send_chat",
        SendChatPayload {
            text: text_value.trim(),
        },
    );
    dispatch_serialized(&mut app, &intent, None);
    app.window.entry_chat().set_text("");
}

fn dispatch_serialized<T>(state: &mut AppState, intent: &AppIntent<T>, fallback_error: Option<&str>)
where
    T: serde::Serialize,
{
    let Some(json) = to_json(intent) else {
        show_banner(&state.window, "Falha ao serializar ação.");
        return;
    };
    let Some(core) = state.core.clone() else {
        show_banner(&state.window, text(state.locale, "missing-core"));
        return;
    };
    match core.dispatch(&json) {
        Ok(Some(response)) => {
            let message = parse_runtime_message(&response)
                .or(fallback_error.map(str::to_string))
                .unwrap_or(response);
            show_banner(&state.window, &message);
            set_status(&state.window, &message);
        }
        Ok(None) => {
            hide_banner(&state.window);
        }
        Err(err) => {
            let message = fallback_error
                .map(str::to_string)
                .unwrap_or_else(|| err.to_string());
            show_banner(&state.window, &message);
            set_status(&state.window, &message);
        }
    }
}

fn process_event(app: &mut AppState, ev: &AppEvent) {
    let _ = ev.sequence;
    let _ = &ev.timestamp;
    if let Some(text_value) = ev.text() {
        let lbl = gtk::Label::new(Some(&text_value));
        lbl.set_halign(gtk::Align::Start);
        lbl.set_wrap(true);
        app.window.list_chat().append(&lbl);
        if ev.kind == "error" || ev.kind == "system" {
            push_toast(&app.window, &text_value);
        }
        if let Some(adj) = app
            .window
            .list_chat()
            .parent()
            .and_then(|p| p.parent())
            .and_then(|p| p.downcast::<gtk::ScrolledWindow>().ok())
            .map(|sw| sw.vadjustment())
        {
            glib::timeout_add_local_once(
                std::time::Duration::from_millis(10),
                glib::clone!(
                    #[weak]
                    adj,
                    move || {
                        adj.set_value(adj.upper() - adj.page_size());
                    }
                ),
            );
        }
    }
}

fn load_css() {
    let provider = gtk::CssProvider::new();
    provider.load_from_data(include_str!("../style.css"));
    gtk::style_context_add_provider_for_display(
        &gtk::gdk::Display::default().expect("Could not connect to a display."),
        &provider,
        gtk::STYLE_PROVIDER_PRIORITY_APPLICATION,
    );
}

fn disable_session_actions(window: &window::TrucoWindow) {
    window.btn_start_demo().set_sensitive(false);
    window.btn_host_online().set_sensitive(false);
    window.btn_join_online().set_sensitive(false);
}

fn apply_locale(window: &window::TrucoWindow, locale: Locale) {
    window.set_title(Some(text(locale, "app-title")));
    window
        .lbl_lobby_title()
        .set_label(text(locale, "app-title"));
    window
        .lbl_lobby_subtitle()
        .set_label(text(locale, "app-subtitle"));
    window
        .btn_start_demo()
        .set_label(text(locale, "play-offline"));
    window
        .btn_host_online()
        .set_label(text(locale, "create-room"));
    window
        .btn_join_online()
        .set_label(text(locale, "join-room"));
    window.btn_copy_invite().set_label(text(locale, "copy-key"));
    window
        .btn_start_online_match()
        .set_label(text(locale, "start-match"));
    window
        .btn_leave_online()
        .set_label(text(locale, "leave-room"));
    window.btn_send_chat().set_label(text(locale, "send"));
}

fn show_banner(window: &window::TrucoWindow, message: &str) {
    window.lbl_banner().set_label(message);
    window.banner_revealer().set_reveal_child(true);
}

fn hide_banner(window: &window::TrucoWindow) {
    window.banner_revealer().set_reveal_child(false);
}

fn push_toast(window: &window::TrucoWindow, message: &str) {
    let toast = adw::Toast::new(message);
    toast.set_timeout(3);
    window.toast_overlay().add_toast(toast);
}

fn set_status(window: &window::TrucoWindow, text_value: &str) {
    window.lbl_status_chip().set_label(text_value);
}

fn fallback_name(name: &str, locale: Locale) -> String {
    if name.trim().is_empty() {
        if locale == Locale::EnUs {
            "You".to_string()
        } else {
            "Voce".to_string()
        }
    } else {
        name.trim().to_string()
    }
}

fn parse_runtime_message(response: &str) -> Option<String> {
    let value: serde_json::Value = serde_json::from_str(response).ok()?;
    value
        .get("message")
        .and_then(|m| m.as_str())
        .map(str::to_string)
}

fn dispatch_game_action(core: &TrucoCore, action: &'static str, card_index: Option<usize>) {
    let intent = AppIntent::with_payload("game_action", GameActionPayload { action, card_index });
    if let Some(json) = to_json(&intent) {
        let _ = core.dispatch(&json);
    }
}

fn clear_box(bx: &gtk::Box) {
    while let Some(child) = bx.first_child() {
        bx.remove(&child);
    }
}

fn clear_listbox(lb: &gtk::ListBox) {
    while let Some(child) = lb.first_child() {
        lb.remove(&child);
    }
}

fn update_ui(
    window: &window::TrucoWindow,
    bundle: &SnapshotBundle,
    core: &TrucoCore,
    locale: Locale,
) {
    let mode = bundle.mode.as_deref().unwrap_or("idle");
    let status_text = bundle
        .connection
        .as_ref()
        .and_then(|c| c.status.as_ref())
        .cloned()
        .unwrap_or_else(|| mode.to_string());
    set_status(window, &status_text.replace('_', " "));

    if mode == "offline_match"
        || mode == "host_match"
        || mode == "client_match"
        || mode == "match_over"
    {
        window.main_stack().set_visible_child_name("game");
        if let Some(ref snap) = bundle.game {
            update_game_ui(window, snap, core, locale);
        }
    } else if mode == "host_lobby" || mode == "client_lobby" {
        window.main_stack().set_visible_child_name("online_lobby");
        update_lobby_ui(window, bundle, core, locale, mode);
    } else {
        window.main_stack().set_visible_child_name("lobby");
    }
}

fn update_lobby_ui(
    window: &window::TrucoWindow,
    bundle: &SnapshotBundle,
    core: &TrucoCore,
    locale: Locale,
    mode: &str,
) {
    window
        .lbl_online_status()
        .set_label(if mode == "host_lobby" {
            text(locale, "online-room")
        } else {
            "Connected"
        });

    if let Some(lobby) = &bundle.lobby {
        if let Some(key) = &lobby.invite_key {
            window
                .lbl_invite_key_display()
                .set_label(&format!("Chave: {key}"));
            window.btn_copy_invite().set_sensitive(true);
        } else {
            window
                .lbl_invite_key_display()
                .set_label("Chave: (Convidado)");
            window.btn_copy_invite().set_sensitive(false);
        }

        clear_listbox(&window.list_slots());
        if let Some(slots) = &lobby.slots {
            for (i, name) in slots.iter().enumerate() {
                let row = gtk::Box::new(gtk::Orientation::Horizontal, 8);
                row.set_halign(gtk::Align::Fill);
                row.set_hexpand(true);
                let connected = lobby
                    .connected_seats
                    .as_ref()
                    .and_then(|m| m.get(&i.to_string()).copied())
                    .unwrap_or(false);
                let mut display = if name.is_empty() {
                    "Aguardando...".to_string()
                } else {
                    name.clone()
                };
                if i as i32 == lobby.host_seat.unwrap_or(0) {
                    display.push_str(" [host]");
                }
                if !name.is_empty() && !connected {
                    display.push_str(" [offline]");
                }
                let lbl = gtk::Label::new(Some(&display));
                lbl.set_hexpand(true);
                lbl.set_xalign(0.0);
                row.append(&lbl);

                if Some(i as i32) == lobby.assigned_seat {
                    let me_lbl = gtk::Label::new(Some(text(locale, "you")));
                    me_lbl.add_css_class("ladder-active");
                    row.append(&me_lbl);
                } else if name.is_empty() && mode.contains("host") {
                    let btn_invite = gtk::Button::with_label("Convite");
                    btn_invite.add_css_class("pill-button");
                    let core_inv = core.clone();
                    btn_invite.connect_clicked(move |_| {
                        let intent = AppIntent::with_payload(
                            "request_replacement_invite",
                            ReplacementInvitePayload { target_seat: i },
                        );
                        if let Some(json) = to_json(&intent) {
                            let _ = core_inv.dispatch(&json);
                        }
                    });
                    row.append(&btn_invite);
                } else if !name.is_empty() {
                    let btn_vote = gtk::Button::with_label("Votar Host");
                    btn_vote.add_css_class("pill-button");
                    let core_vote = core.clone();
                    btn_vote.connect_clicked(move |_| {
                        let intent = AppIntent::with_payload(
                            "vote_host",
                            HostVotePayload { candidate_seat: i },
                        );
                        if let Some(json) = to_json(&intent) {
                            let _ = core_vote.dispatch(&json);
                        }
                    });
                    row.append(&btn_vote);
                }

                let lb_row = gtk::ListBoxRow::new();
                lb_row.set_child(Some(&row));
                window.list_slots().append(&lb_row);
            }
        }
    }

    window
        .btn_start_online_match()
        .set_visible(mode == "host_lobby");
}

fn update_game_ui(
    window: &window::TrucoWindow,
    snapshot: &GameSnapshot,
    core: &TrucoCore,
    locale: Locale,
) {
    let local_idx = snapshot.current_player_idx.unwrap_or(0);
    let my_team = snapshot
        .players
        .as_ref()
        .and_then(|p| p.iter().find(|pl| pl.id == local_idx))
        .map(|pl| pl.team)
        .unwrap_or(0);
    let n = snapshot.num_players.unwrap_or(2) as usize;

    if snapshot.match_finished == Some(true) {
        window.game_over_overlay().set_visible(true);
        if snapshot.winner_team == Some(my_team) {
            window.lbl_winner().set_label(if locale == Locale::EnUs {
                "VICTORY!"
            } else {
                "VITORIA!"
            });
        } else {
            window.lbl_winner().set_label(if locale == Locale::EnUs {
                "DEFEAT"
            } else {
                "DERROTA"
            });
        }
    } else {
        window.game_over_overlay().set_visible(false);
    }

    let hud_box = window.hud_box();
    clear_box(&hud_box);

    let log_box = window.log_box();
    clear_box(&log_box);
    if let Some(logs) = &snapshot.logs {
        let recent = if logs.len() > 8 {
            &logs[logs.len() - 8..]
        } else {
            logs.as_slice()
        };
        for entry in recent {
            let lbl = gtk::Label::new(Some(entry));
            lbl.add_css_class("log-entry");
            lbl.set_halign(gtk::Align::End);
            log_box.append(&lbl);
        }
    }

    let us = snapshot
        .match_points
        .as_ref()
        .and_then(|p| p.get("0").copied())
        .unwrap_or(0);
    let them = snapshot
        .match_points
        .as_ref()
        .and_then(|p| p.get("1").copied())
        .unwrap_or(0);
    let stake = snapshot
        .current_hand
        .as_ref()
        .and_then(|h| h.stake)
        .unwrap_or(1);
    let round = snapshot
        .current_hand
        .as_ref()
        .and_then(|h| h.round)
        .unwrap_or(1);

    let score_us = gtk::Label::new(None);
    score_us.set_markup(&format!(
        "<span size='large'><b>{}</b></span>\n<span size='28000'><b>{}</b></span>",
        if locale == Locale::EnUs { "US" } else { "NOS" },
        us
    ));
    score_us.set_justify(gtk::Justification::Center);
    score_us.add_css_class("score-hud");

    let mid_box = gtk::Box::new(gtk::Orientation::Vertical, 4);
    mid_box.set_halign(gtk::Align::Center);

    let stake_lbl = gtk::Label::new(None);
    stake_lbl.set_markup(&format!(
        "<span size='large'><b>{}</b></span>\n<span size='28000' color='#ffcc00'><b>{}</b></span>",
        if locale == Locale::EnUs {
            "STAKE"
        } else {
            "VALE"
        },
        stake
    ));
    stake_lbl.set_justify(gtk::Justification::Center);
    stake_lbl.add_css_class("stake-badge");
    mid_box.append(&stake_lbl);

    let ladder = gtk::Box::new(gtk::Orientation::Horizontal, 4);
    ladder.set_halign(gtk::Align::Center);
    for &level in &[1, 3, 6, 9, 12] {
        let label = match level {
            1 => "1",
            3 => "T",
            6 => "6",
            9 => "9",
            _ => "12",
        };
        let l = gtk::Label::new(Some(label));
        l.add_css_class(if stake >= level {
            "ladder-active"
        } else {
            "ladder-dim"
        });
        ladder.append(&l);
    }
    mid_box.append(&ladder);

    let round_lbl = gtk::Label::new(Some(&format!("{round}/3")));
    round_lbl.add_css_class("turn-pill");
    mid_box.append(&round_lbl);

    if let Some(tp) = snapshot.turn_player {
        if let Some(players) = &snapshot.players {
            if let Some(turn_p) = players.iter().find(|p| p.id == tp) {
                let turn_lbl = gtk::Label::new(Some(&format!(
                    "{} {}",
                    if locale == Locale::EnUs {
                        "Turn:"
                    } else {
                        "Vez:"
                    },
                    turn_p.name
                )));
                turn_lbl.add_css_class("turn-pill");
                mid_box.append(&turn_lbl);
            }
        }
    }

    let score_them = gtk::Label::new(None);
    score_them.set_markup(&format!(
        "<span size='large'><b>{}</b></span>\n<span size='28000'><b>{}</b></span>",
        if locale == Locale::EnUs {
            "THEM"
        } else {
            "ELES"
        },
        them
    ));
    score_them.set_justify(gtk::Justification::Center);
    score_them.add_css_class("score-hud");

    hud_box.append(&score_us);
    hud_box.append(&mid_box);
    hud_box.append(&score_them);

    let opp_box = window.opponent_box();
    clear_box(&opp_box);
    let left_box = window.left_player_box();
    clear_box(&left_box);
    let right_box = window.right_player_box();
    clear_box(&right_box);

    if let Some(players) = &snapshot.players {
        let top_offset = if n == 4 { 2 } else { 1 };
        let top_id = (local_idx + top_offset) % n as i32;
        if let Some(top) = players.iter().find(|p| p.id == top_id) {
            render_opponent(
                &opp_box,
                top,
                snapshot.turn_player == Some(top.id),
                false,
                snapshot,
            );
        }

        if n == 4 {
            let right_id = (local_idx + 1) % 4;
            if let Some(right) = players.iter().find(|p| p.id == right_id) {
                render_opponent(
                    &right_box,
                    right,
                    snapshot.turn_player == Some(right.id),
                    true,
                    snapshot,
                );
            }
            let left_id = (local_idx + 3) % 4;
            if let Some(left) = players.iter().find(|p| p.id == left_id) {
                render_opponent(
                    &left_box,
                    left,
                    snapshot.turn_player == Some(left.id),
                    true,
                    snapshot,
                );
            }
        }
    }

    let center_box = window.center_box();
    clear_box(&center_box);
    if let Some(hand) = &snapshot.current_hand {
        let vira_stack = gtk::Box::new(gtk::Orientation::Vertical, 8);
        vira_stack.append(&gtk::Label::new(Some("VIRA")));
        if let Some(vira) = &hand.vira {
            vira_stack.append(&create_card_widget(Some(vira)));
        }
        center_box.append(&vira_stack);

        let played_box = gtk::Box::new(gtk::Orientation::Horizontal, -30);
        played_box.set_size_request(180, 160);
        played_box.set_valign(gtk::Align::Center);
        played_box.set_halign(gtk::Align::Center);
        if let Some(rounds) = &hand.round_cards {
            for playing in rounds {
                let col = gtk::Box::new(gtk::Orientation::Vertical, 4);
                let pname = snapshot
                    .players
                    .as_ref()
                    .and_then(|p| p.iter().find(|pl| pl.id == playing.player_id))
                    .map(|pl| pl.name.as_str())
                    .unwrap_or("?");
                col.append(&gtk::Label::new(Some(pname)));
                col.append(&create_card_widget(Some(&playing.card)));
                played_box.append(&col);
            }
        }
        center_box.append(&played_box);

        let manilha_stack = gtk::Box::new(gtk::Orientation::Vertical, 8);
        manilha_stack.append(&gtk::Label::new(Some("MANILHA")));
        if let Some(m) = &hand.manilha {
            let m_lbl = gtk::Label::new(None);
            m_lbl.set_size_request(86, 124);
            m_lbl.set_markup(&format!("<span size='34000'><b>{m}</b></span>"));
            m_lbl.add_css_class("manilha-area");
            m_lbl.add_css_class("manilha-glow");
            manilha_stack.append(&m_lbl);
        }
        center_box.append(&manilha_stack);
    }

    let bottom_box = window.bottom_box();
    clear_box(&bottom_box);

    let pending_for = snapshot.pending_raise_for.unwrap_or(-1);
    let can_ask = snapshot.can_ask_truco.unwrap_or(false);
    let is_my_turn = snapshot.turn_player == Some(local_idx);
    let match_over = snapshot.match_finished.unwrap_or(false);

    if !match_over {
        let action_box = gtk::Box::new(gtk::Orientation::Horizontal, 16);
        action_box.set_halign(gtk::Align::Center);

        if pending_for == my_team {
            let btn_accept = gtk::Button::with_label(if locale == Locale::EnUs {
                "ACCEPT"
            } else {
                "ACEITAR"
            });
            btn_accept.add_css_class("btn-accept");
            let core_a = core.clone();
            btn_accept.connect_clicked(move |_| dispatch_game_action(&core_a, "accept", None));
            action_box.append(&btn_accept);

            let current_stake = snapshot
                .current_hand
                .as_ref()
                .and_then(|h| h.stake)
                .unwrap_or(1);
            if current_stake < 9 {
                let btn_raise =
                    gtk::Button::with_label(&raise_label_for(next_stake(current_stake)));
                btn_raise.add_css_class("btn-truco");
                let core_r = core.clone();
                btn_raise.connect_clicked(move |_| dispatch_game_action(&core_r, "truco", None));
                action_box.append(&btn_raise);
            }

            let btn_refuse = gtk::Button::with_label(if locale == Locale::EnUs {
                "FOLD"
            } else {
                "CORRER"
            });
            btn_refuse.add_css_class("btn-refuse");
            let core_f = core.clone();
            btn_refuse.connect_clicked(move |_| dispatch_game_action(&core_f, "refuse", None));
            action_box.append(&btn_refuse);
        } else if is_my_turn && can_ask {
            let current_stake = snapshot
                .current_hand
                .as_ref()
                .and_then(|h| h.stake)
                .unwrap_or(1);
            let btn_truco = gtk::Button::with_label(&raise_label_for(next_stake(current_stake)));
            btn_truco.add_css_class("btn-truco");
            let core_t = core.clone();
            btn_truco.connect_clicked(move |_| dispatch_game_action(&core_t, "truco", None));
            action_box.append(&btn_truco);
        }

        bottom_box.append(&action_box);
    }

    if is_my_turn && !match_over {
        let lbl = gtk::Label::new(Some(if locale == Locale::EnUs {
            "YOUR TURN"
        } else {
            "SUA VEZ"
        }));
        lbl.add_css_class("turn-pill");
        bottom_box.append(&lbl);
    }

    if let Some(players) = &snapshot.players {
        if let Some(me) = players.iter().find(|p| p.id == local_idx) {
            let name_row = gtk::Box::new(gtk::Orientation::Horizontal, 8);
            name_row.set_halign(gtk::Align::Center);
            name_row.append(&gtk::Label::new(Some(&me.name.to_uppercase())));
            let team_lbl = gtk::Label::new(Some(&format!("Team {}", me.team + 1)));
            team_lbl.add_css_class(if me.team == 0 {
                "team-badge-us"
            } else {
                "team-badge-them"
            });
            name_row.append(&team_lbl);
            bottom_box.append(&name_row);

            let my_hand = gtk::Box::new(gtk::Orientation::Horizontal, 16);
            if let Some(cards) = &me.hand {
                for (idx, c) in cards.iter().enumerate() {
                    let card_widget = create_card_widget(Some(c));
                    card_widget.add_css_class("card-clickable");
                    let gesture = gtk::GestureClick::new();
                    let core_clone = core.clone();
                    gesture.connect_pressed(move |_, _, _, _| {
                        dispatch_game_action(&core_clone, "play", Some(idx));
                    });
                    card_widget.add_controller(gesture);
                    my_hand.append(&card_widget);
                }
            }
            bottom_box.append(&my_hand);
        }
    }
}

fn render_opponent(
    container: &gtk::Box,
    player: &models::Player,
    is_turn: bool,
    vertical_cards: bool,
    snap: &GameSnapshot,
) {
    let badge = role_badge_for(player.id, snap);
    let mut name_text = String::new();
    if let Some(b) = &badge {
        name_text.push_str(b);
        name_text.push(' ');
    }
    name_text.push_str(&player.name.to_uppercase());
    if player.cpu == Some(true) {
        if is_turn {
            name_text.push_str(" ⟳");
        } else {
            name_text.push_str(if player.provisional_cpu == Some(true) {
                " 🤖*"
            } else {
                " 🤖"
            });
        }
    }
    if is_turn && player.cpu != Some(true) {
        name_text.push_str(" ◀");
    }
    let name_lbl = gtk::Label::new(Some(&name_text));
    name_lbl.add_css_class(if is_turn {
        "opponent-pill-active"
    } else {
        "opponent-pill"
    });
    container.append(&name_lbl);

    let team_lbl = gtk::Label::new(Some(&format!("Team {}", player.team + 1)));
    team_lbl.add_css_class(if player.team == 0 {
        "team-badge-us"
    } else {
        "team-badge-them"
    });
    container.append(&team_lbl);

    let count = player.hand.as_ref().map(|h| h.len()).unwrap_or(3);
    let hand_box = gtk::Box::new(
        if vertical_cards {
            gtk::Orientation::Vertical
        } else {
            gtk::Orientation::Horizontal
        },
        if vertical_cards { -30 } else { -20 },
    );
    for _ in 0..count {
        hand_box.append(&create_card_widget(None));
    }
    container.append(&hand_box);
}

fn role_badge_for(player_id: i32, snap: &GameSnapshot) -> Option<String> {
    let hand = snap.current_hand.as_ref()?;
    let n = snap.num_players.unwrap_or(2);
    let dealer = hand.dealer?;
    if player_id == dealer {
        return Some("🃏".to_string());
    }
    if player_id == (dealer + 1) % n {
        return Some("✋".to_string());
    }
    if n == 4 && player_id == (dealer + n - 1) % n {
        return Some("🦶".to_string());
    }
    None
}

fn next_stake(s: i32) -> i32 {
    match s {
        1 => 3,
        3 => 6,
        6 => 9,
        9 => 12,
        _ => s,
    }
}

fn raise_label_for(stake: i32) -> String {
    match stake {
        3 => "TRUCO!",
        6 => "SEIS!",
        9 => "NOVE!",
        12 => "DOZE!",
        _ => "TRUCO!",
    }
    .to_string()
}

fn create_card_widget(card_opt: Option<&models::Card>) -> gtk::Box {
    let container = gtk::Box::new(gtk::Orientation::Vertical, 0);
    container.set_size_request(86, 124);

    if let Some(card) = card_opt {
        container.add_css_class("card");
        if card.is_red() {
            container.add_css_class("card-red");
        }

        let top_str = format!(
            "<span size='large'><b>{}</b></span>\n{}",
            card.rank,
            card.suit_symbol()
        );
        let top_lbl = gtk::Label::new(None);
        top_lbl.set_markup(&top_str);
        top_lbl.set_halign(gtk::Align::Start);
        top_lbl.set_valign(gtk::Align::Start);
        top_lbl.set_margin_top(8);
        top_lbl.set_margin_start(8);

        let center_lbl = gtk::Label::new(None);
        center_lbl.set_markup(&format!(
            "<span size='32000'><b>{}</b></span>",
            card.suit_symbol()
        ));
        center_lbl.set_vexpand(true);

        let bottom_lbl = gtk::Label::new(None);
        bottom_lbl.set_markup(&top_str);
        bottom_lbl.set_halign(gtk::Align::End);
        bottom_lbl.set_valign(gtk::Align::End);
        bottom_lbl.set_margin_bottom(8);
        bottom_lbl.set_margin_end(8);

        container.append(&top_lbl);
        container.append(&center_lbl);
        container.append(&bottom_lbl);
    } else {
        container.add_css_class("card-back");
    }

    container
}
