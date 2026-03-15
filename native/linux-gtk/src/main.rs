mod truco_core;
mod models;
mod window;

use std::rc::Rc;
use std::cell::RefCell;
use adw::prelude::*;
use gtk::glib;
use truco_core::TrucoCore;

struct AppState {
    pub core: TrucoCore,
    pub window: window::TrucoWindow,
    pub last_snapshot_str: String,
    pub last_pending_raise: Option<i32>,
    pub truco_flash_ticks: u32,
    pub last_trick_seq_viewed: i32,
    pub event_history: Vec<String>,
}

fn main() {
    let app = adw::Application::builder()
        .application_id("dev.truco.Native")
        .build();

    app.connect_startup(|_| {
        load_css();
    });

    app.connect_activate(|app| {
        let core = TrucoCore::new();
        let window = window::TrucoWindow::new(app);

        let state = Rc::new(RefCell::new(AppState {
            core: core.clone(),
            window: window.clone(),
            last_snapshot_str: String::new(),
            last_pending_raise: None,
            truco_flash_ticks: 0,
            last_trick_seq_viewed: -1,
            event_history: Vec::new(),
        }));

        // Start button — read setup form values
        let btn_start = window.btn_start_demo();
        let win_start = window.clone();
        let core_start = core.clone();
        btn_start.connect_clicked(move |_| {
            let player_name = win_start.entry_player_name().text().to_string();
            let name = if player_name.is_empty() { "Voce".to_string() } else { player_name };
            
            let num_idx = win_start.dd_num_players().selected();
            let num_players = if num_idx == 1 { 4 } else { 2 };
            
            // Set locale
            let locale_idx = win_start.dd_locale().selected();
            let locale = if locale_idx == 1 { "en-US" } else { "pt-BR" };
            let _ = core_start.dispatch(&format!(r#"{{"kind":"set_locale","payload":{{"locale":"{}"}}}}"#, locale));
            
            // Build player names and CPU flags
            let (names, cpus) = if num_players == 4 {
                (
                    format!(r#"["{}","CPU-Direita","CPU-Parceiro","CPU-Esquerda"]"#, name),
                    "[false,true,true,true]".to_string(),
                )
            } else {
                (
                    format!(r#"["{}","CPU-Oponente"]"#, name),
                    "[false,true]".to_string(),
                )
            };
            
            let intent = format!(
                r#"{{"kind":"new_offline_game","payload":{{"player_names":{},"cpu_flags":{}}}}}"#,
                names, cpus
            );
            let _ = core_start.dispatch(&intent);
        });

        let btn_back = window.btn_back_lobby();
        let win_back = window.clone();
        btn_back.connect_clicked(move |_| {
            win_back.game_over_overlay().set_visible(false);
            win_back.main_stack().set_visible_child_name("lobby");
        });

        // Leave Match Button
        let btn_leave_match = window.btn_leave_match();
        let win_leave_match = window.clone();
        let core_leave_match = core.clone();
        btn_leave_match.connect_clicked(move |_| {
            let _ = core_leave_match.dispatch(r#"{"kind":"close_session"}"#);
            win_leave_match.game_over_overlay().set_visible(false);
            win_leave_match.trick_end_overlay().set_visible(false);
            clear_listbox(&win_leave_match.list_chat());
        });

        // Online Host button
        let btn_host_online = window.btn_host_online();
        let win_host = window.clone();
        let core_host = core.clone();
        btn_host_online.connect_clicked(move |_| {
            let player_name = win_host.entry_player_name().text().to_string();
            let name = if player_name.is_empty() { "Voce".to_string() } else { player_name };
            let num_idx = win_host.dd_num_players().selected();
            let num_players = if num_idx == 1 { 4 } else { 2 };
            let relay_url = win_host.entry_relay_url().text().to_string();
            let intent = if relay_url.trim().is_empty() {
                format!(
                    r#"{{"kind":"create_host_session","payload":{{"host_name":"{}","num_players":{}}}}}"#,
                    name, num_players
                )
            } else {
                let safe_relay = relay_url.replace("\"", "\\\"");
                format!(
                    r#"{{"kind":"create_host_session","payload":{{"host_name":"{}","num_players":{},"relay_url":"{}"}}}}"#,
                    name, num_players, safe_relay
                )
            };
            let _ = core_host.dispatch(&intent);
        });

        // Online Join button
        let btn_join_online = window.btn_join_online();
        let win_join = window.clone();
        let core_join = core.clone();
        btn_join_online.connect_clicked(move |_| {
            let player_name = win_join.entry_player_name().text().to_string();
            let name = if player_name.is_empty() { "Voce".to_string() } else { player_name };
            let key = win_join.entry_invite_key().text().to_string();
            if key.is_empty() { return; }
            let desired_role = match win_join.dd_desired_role().selected() {
                1 => "partner",
                2 => "opponent",
                _ => "auto",
            };
            let intent = format!(
                r#"{{"kind":"join_session","payload":{{"key":"{}","player_name":"{}","desired_role":"{}"}}}}"#,
                key, name, desired_role
            );
            let _ = core_join.dispatch(&intent);
        });

        // Start Match Online (Host only)
        let btn_start_online = window.btn_start_online_match();
        let core_start_online = core.clone();
        btn_start_online.connect_clicked(move |_| {
            let _ = core_start_online.dispatch(r#"{"kind":"start_hosted_match"}"#);
        });

        // Leave Online
        let btn_leave = window.btn_leave_online();
        let win_leave = window.clone();
        let core_leave = core.clone();
        btn_leave.connect_clicked(move |_| {
            let _ = core_leave.dispatch(r#"{"kind":"close_session"}"#);
            clear_listbox(&win_leave.list_chat());
        });

        // Send Chat
        let btn_send_chat = window.btn_send_chat();
        let win_chat = window.clone();
        let core_chat = core.clone();
        btn_send_chat.connect_clicked(move |_| {
            let text = win_chat.entry_chat().text().to_string();
            if !text.is_empty() {
                let safe_text = text.replace("\"", "\\\"");
                let intent = format!(r#"{{"kind":"send_chat","payload":{{"text":"{}"}}}}"#, safe_text);
                let _ = core_chat.dispatch(&intent);
                win_chat.entry_chat().set_text("");
            }
        });

        let entry_chat = window.entry_chat();
        let win_chat2 = window.clone();
        let core_chat2 = core.clone();
        entry_chat.connect_activate(move |_| {
            let text = win_chat2.entry_chat().text().to_string();
            if !text.is_empty() {
                let safe_text = text.replace("\"", "\\\"");
                let intent = format!(r#"{{"kind":"send_chat","payload":{{"text":"{}"}}}}"#, safe_text);
                let _ = core_chat2.dispatch(&intent);
                win_chat2.entry_chat().set_text("");
            }
        });

        // Game loop polling
        glib::timeout_add_local(std::time::Duration::from_millis(50), glib::clone!(#[strong] state, move || {
            let mut s = state.borrow_mut();
            if let Some(event_str) = s.core.poll_event() {
                if let Ok(ev) = serde_json::from_str::<models::AppEvent>(&event_str) {
                    if ev.kind == "chat" || ev.kind == "system" || ev.kind == "replacement_invite" {
                        let text = if ev.kind == "chat" {
                            let author = ev.payload.as_ref().and_then(|p| p.get("author").and_then(|a| a.as_str())).unwrap_or("?");
                            let msg = ev.payload.as_ref().and_then(|p| p.get("text").and_then(|t| t.as_str())).unwrap_or("");
                            format!("{}: {}", author, msg)
                        } else if ev.kind == "system" {
                            ev.payload.as_ref().and_then(|p| p.get("text").and_then(|t| t.as_str())).unwrap_or("").to_string()
                        } else {
                            let link = ev.payload.as_ref().and_then(|p| p.get("invite_key").and_then(|t| t.as_str())).unwrap_or("");
                            format!("Link de substituição: {}", link)
                        };
                        let lbl = gtk::Label::new(Some(&text));
                        lbl.set_halign(gtk::Align::Start);
                        lbl.set_wrap(true);
                        s.window.list_chat().append(&lbl);
                        s.event_history.push(text);
                        if s.event_history.len() > 40 {
                            let drain_to = s.event_history.len() - 40;
                            s.event_history.drain(..drain_to);
                        }
                        
                        // scroll to bottom
                        if let Some(adj) = s.window.list_chat().parent().and_then(|p| p.parent()).and_then(|p| p.downcast::<gtk::ScrolledWindow>().ok()).map(|sw| sw.vadjustment()) {
                            glib::timeout_add_local_once(std::time::Duration::from_millis(10), glib::clone!(#[weak] adj, move || {
                                adj.set_value(adj.upper() - adj.page_size());
                            }));
                        }
                    }
                }
            }
            if let Some(snap_str) = s.core.snapshot() {
                if snap_str != s.last_snapshot_str {
                    s.last_snapshot_str = snap_str.clone();
                    if let Some(bundle) = models::SnapshotBundle::from_json(&snap_str) {
                         let AppState { window, core, event_history, last_trick_seq_viewed, .. } = &mut *s;
                         update_ui(window, &bundle, core, event_history, last_trick_seq_viewed);
                    }
                }
            }
            glib::ControlFlow::Continue
        }));

        adw::prelude::GtkWindowExt::present(&window);
    });

    app.run();
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

fn dispatch_game_action(core: &TrucoCore, action: &str, card_index: Option<usize>) {
    let json = if let Some(idx) = card_index {
        format!(r#"{{"kind":"game_action","payload":{{"action":"{}","card_index":{}}}}}"#, action, idx)
    } else {
        format!(r#"{{"kind":"game_action","payload":{{"action":"{}"}}}}"#, action)
    };
    let _ = core.dispatch(&json);
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

fn update_ui(window: &window::TrucoWindow, bundle: &models::SnapshotBundle, core: &TrucoCore, event_history: &[String], last_trick_seq_viewed: &mut i32) {
    let mode = bundle.mode.as_deref().unwrap_or("idle");
    
    if mode == "offline_match" || mode == "host_match" || mode == "client_match" || mode == "match_over" {
        window.main_stack().set_visible_child_name("game");
        if let Some(ref snap) = bundle.match_snapshot {
            update_game_ui(window, snap, bundle, core, event_history, last_trick_seq_viewed);
        }
    } else if mode == "host_lobby" || mode == "client_lobby" {
        window.main_stack().set_visible_child_name("online_lobby");
        update_lobby_ui(window, bundle, core, mode);
    } else {
        window.main_stack().set_visible_child_name("lobby");
    }
}

fn update_lobby_ui(window: &window::TrucoWindow, bundle: &models::SnapshotBundle, core: &TrucoCore, mode: &str) {
    let status = bundle.connection.as_ref()
        .and_then(|v| v.get("status").and_then(|s| s.as_str()))
        .unwrap_or(mode);
    let online = bundle.connection.as_ref()
        .and_then(|v| v.get("is_online").and_then(|s| s.as_bool()))
        .unwrap_or(true);
    let backlog = bundle.diagnostics.as_ref()
        .and_then(|v| v.get("event_backlog").and_then(|n| n.as_i64()))
        .unwrap_or(0);
    
    if let Some(lobby_val) = &bundle.lobby {
        if let Ok(lobby) = serde_json::from_value::<models::LobbySnapshot>(lobby_val.clone()) {
            let header = if let Some(role) = lobby.role.as_deref() {
                format!("{}  •  {}  •  fila {}  •  papel {}", status, if online { "online" } else { "offline" }, backlog, role)
            } else {
                format!("{}  •  {}  •  fila {}", status, if online { "online" } else { "offline" }, backlog)
            };
            window.lbl_online_status().set_label(&header);
            if let Some(key) = &lobby.invite_key {
                window.lbl_invite_key_display().set_label(&format!("Chave: {}", key));
            } else {
                window.lbl_invite_key_display().set_label("Chave: (Convidado)");
            }
            
            clear_listbox(&window.list_slots());
            if let Some(slot_states) = bundle.ui.as_ref().and_then(|u| u.lobby_slots.as_ref()) {
                for slot in slot_states {
                    let row = gtk::Box::new(gtk::Orientation::Horizontal, 8);
                    let name = slot.name.clone().unwrap_or_default();
                    let mut display = if slot.is_empty { "Aguardando...".to_string() } else { name.clone() };
                    if slot.is_host {
                        display.push_str(" [host]");
                    }
                    if slot.is_provisional_cpu {
                        display.push_str(" [cpu]");
                    }
                    if !slot.is_empty && !slot.is_connected {
                        display.push_str(" [offline]");
                    }
                    let lbl = gtk::Label::new(Some(&display));
                    row.append(&lbl);
                    
                    if slot.is_local {
                        let me_lbl = gtk::Label::new(Some("(você)"));
                        me_lbl.add_css_class("ladder-active");
                        row.append(&me_lbl);
                    } else if slot.can_request_replacement {
                        let btn_invite = gtk::Button::with_label("Convite");
                        let core_inv = core.clone();
                        let seat = slot.seat;
                        btn_invite.connect_clicked(move |_| {
                            let _ = core_inv.dispatch(&format!(r#"{{"kind":"request_replacement_invite","payload":{{"target_seat":{}}}}}"#, seat));
                        });
                        row.append(&btn_invite);
                    } else if slot.can_vote_host {
                        let btn_vote = gtk::Button::with_label("Votar Host");
                        let core_vote = core.clone();
                        let seat = slot.seat;
                        btn_vote.connect_clicked(move |_| {
                            let _ = core_vote.dispatch(&format!(r#"{{"kind":"vote_host","payload":{{"candidate_seat":{}}}}}"#, seat));
                        });
                        row.append(&btn_vote);
                    }
                    
                    let lb_row = gtk::ListBoxRow::new();
                    lb_row.set_child(Some(&row));
                    window.list_slots().append(&lb_row);
                }
            } else if let Some(slots) = lobby.slots {
                for (i, name) in slots.iter().enumerate() {
                    let row = gtk::Box::new(gtk::Orientation::Horizontal, 8);
                    let connected = lobby.connected_seats.as_ref()
                        .and_then(|m| m.get(&i.to_string()).copied())
                        .unwrap_or(false);
                    let mut display = if name.is_empty() { "Aguardando...".to_string() } else { name.clone() };
                    if i as i32 == lobby.host_seat.unwrap_or(0) {
                        display.push_str(" [host]");
                    }
                    if !name.is_empty() && !connected {
                        display.push_str(" [offline]");
                    }
                    let lbl = gtk::Label::new(Some(&display));
                    row.append(&lbl);
                    let lb_row = gtk::ListBoxRow::new();
                    lb_row.set_child(Some(&row));
                    window.list_slots().append(&lb_row);
                }
            }
        }
    } else {
        let header = format!("{}  •  {}  •  fila {}", status, if online { "online" } else { "offline" }, backlog);
        window.lbl_online_status().set_label(&header);
    }
    
    window.btn_start_online_match().set_visible(mode == "host_lobby");
}

fn update_game_ui(window: &window::TrucoWindow, snapshot: &models::GameSnapshot, bundle: &models::SnapshotBundle, core: &TrucoCore, event_history: &[String], last_trick_seq_viewed: &mut i32) {
    let mode = bundle.mode.as_deref().unwrap_or("idle");
    
    if mode == "offline_match" || mode == "host_match" || mode == "client_match" || mode == "match_over" {
        window.main_stack().set_visible_child_name("game");
    } else {
        window.main_stack().set_visible_child_name("online_lobby");
        *last_trick_seq_viewed = -1;
        window.trick_end_overlay().set_visible(false);
        return;
    }

    let local_idx = snapshot.current_player_idx.unwrap_or(0);
    let my_team = snapshot.players.as_ref()
        .and_then(|p| p.iter().find(|pl| pl.id == local_idx))
        .map(|pl| pl.team).unwrap_or(0);
    let n = snapshot.num_players.unwrap_or(2) as usize;

    // Trick End Animation Tracker
    if let Some(trick_seq) = snapshot.last_trick_seq {
        if trick_seq > 0 {
            if *last_trick_seq_viewed == -1 {
                *last_trick_seq_viewed = trick_seq;
            } else if trick_seq > *last_trick_seq_viewed {
                *last_trick_seq_viewed = trick_seq;
                
                // Show banner
                let winner_team = snapshot.last_trick_team.unwrap_or(-1);
                let tie = snapshot.last_trick_tie.unwrap_or(false);
                
                let emoji = if tie { "😐" } else if winner_team == my_team { "🎉" } else { "😢" };
                let text = if tie { "EMPATE!" } else if winner_team == my_team { "VOCE VENCEU A VAZA!" } else { "ELES VENCERAM" };
                
                window.lbl_trick_emoji().set_label(emoji);
                window.lbl_trick_text().set_label(text);
                if tie {
                    window.lbl_trick_text().add_css_class("text-tie");
                } else if winner_team == my_team {
                    window.lbl_trick_text().add_css_class("text-victory");
                } else {
                    window.lbl_trick_text().add_css_class("text-defeat");
                }
                
                window.trick_end_overlay().set_visible(true);
                let win_clone = window.clone();
                glib::timeout_add_local_once(std::time::Duration::from_millis(1800), move || {
                    win_clone.trick_end_overlay().set_visible(false);
                });
            }
        }
    }

    // Match Over
    if snapshot.match_finished == Some(true) {
        window.game_over_overlay().set_visible(true);
        if snapshot.winner_team == Some(my_team) {
            window.lbl_winner().set_label("VITÓRIA! 🎉");
        } else {
            window.lbl_winner().set_label("DERROTA 😢");
        }
    } else {
        window.game_over_overlay().set_visible(false);
    }

    // HUD Layer
    let hud_box = window.hud_box();
    clear_box(&hud_box);
    
    // Log Layer
    let log_box = window.log_box();
    clear_box(&log_box);
    let mut merged_logs = snapshot.logs.clone().unwrap_or_default();
    merged_logs.extend(event_history.iter().cloned());
    if !merged_logs.is_empty() {
        let max_logs = 10;
        let recent = if merged_logs.len() > max_logs { &merged_logs[merged_logs.len() - max_logs..] } else { merged_logs.as_slice() };
        for entry in recent {
            let lbl = gtk::Label::new(Some(entry));
            lbl.add_css_class("log-entry");
            lbl.set_halign(gtk::Align::End);
            log_box.append(&lbl);
        }
    }
    
    let us = snapshot.match_points.as_ref().and_then(|p| p.get("0").copied()).unwrap_or(0);
    let them = snapshot.match_points.as_ref().and_then(|p| p.get("1").copied()).unwrap_or(0);
    let stake = snapshot.current_hand.as_ref().and_then(|h| h.stake).unwrap_or(1);
    let round = snapshot.current_hand.as_ref().and_then(|h| h.round).unwrap_or(1);
    
    let score_us = gtk::Label::new(None);
    score_us.set_markup(&format!("<span size='large'><b>NÓS</b></span>\n<span size='28000'><b>{}</b></span>", us));
    score_us.set_justify(gtk::Justification::Center);
    score_us.add_css_class("score-hud");
    
    // Stake + Ladder + Round
    let mid_box = gtk::Box::new(gtk::Orientation::Vertical, 4);
    mid_box.set_halign(gtk::Align::Center);
    
    let stake_lbl = gtk::Label::new(None);
    stake_lbl.set_markup(&format!("<span size='large'><b>VALE</b></span>\n<span size='28000' color='#ffcc00'><b>{}</b></span>", stake));
    stake_lbl.set_justify(gtk::Justification::Center);
    stake_lbl.add_css_class("stake-badge");
    mid_box.append(&stake_lbl);
    
    // Stake ladder
    let ladder = gtk::Box::new(gtk::Orientation::Horizontal, 4);
    ladder.set_halign(gtk::Align::Center);
    for &level in &[1, 3, 6, 9, 12] {
        let label = match level { 1 => "1", 3 => "T", 6 => "6", 9 => "9", _ => "12" };
        let l = gtk::Label::new(Some(label));
        l.add_css_class(if stake >= level { "ladder-active" } else { "ladder-dim" });
        ladder.append(&l);
    }
    mid_box.append(&ladder);
    
    let round_lbl = gtk::Label::new(Some(&format!("R{}/3", round)));
    round_lbl.add_css_class("turn-pill");
    mid_box.append(&round_lbl);
    
    // Turn indicator with CPU spinner
    if let Some(tp) = snapshot.turn_player {
        if let Some(players) = &snapshot.players {
            if let Some(turn_p) = players.iter().find(|p| p.id == tp) {
                let turn_text = if turn_p.cpu == Some(true) {
                    format!("⟳ Vez: {} (CPU)", turn_p.name)
                } else {
                    format!("Vez: {}", turn_p.name)
                };
                let turn_lbl = gtk::Label::new(Some(&turn_text));
                turn_lbl.add_css_class("turn-pill");
                mid_box.append(&turn_lbl);
            }
        }
    }
    
    let score_them = gtk::Label::new(None);
    score_them.set_markup(&format!("<span size='large'><b>ELES</b></span>\n<span size='28000'><b>{}</b></span>", them));
    score_them.set_justify(gtk::Justification::Center);
    score_them.add_css_class("score-hud");
    
    hud_box.append(&score_us);
    hud_box.append(&mid_box);
    hud_box.append(&score_them);
    
    // Opponents
    let opp_box = window.opponent_box();
    clear_box(&opp_box);
    let left_box = window.left_player_box();
    clear_box(&left_box);
    let right_box = window.right_player_box();
    clear_box(&right_box);
    
    if let Some(players) = &snapshot.players {
        // Top opponent (offset 2 for 4p, offset 1 for 2p)
        let top_offset = if n == 4 { 2 } else { 1 };
        let top_id = (local_idx + top_offset) % n as i32;
        if let Some(top) = players.iter().find(|p| p.id == top_id) {
            render_opponent(&opp_box, top, snapshot.turn_player == Some(top.id), false, snapshot);
        }
        
        if n == 4 {
            // Right (offset 1)
            let right_id = (local_idx + 1) % 4;
            if let Some(right) = players.iter().find(|p| p.id == right_id) {
                render_opponent(&right_box, right, snapshot.turn_player == Some(right.id), true, snapshot);
            }
            // Left (offset 3)
            let left_id = (local_idx + 3) % 4;
            if let Some(left) = players.iter().find(|p| p.id == left_id) {
                render_opponent(&left_box, left, snapshot.turn_player == Some(left.id), true, snapshot);
            }
        }
    }
    
    // Center Table
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
        
        // Find winning card formatted ID
        let winning_card_id = hand.winning_card_id();

        if let Some(rounds) = &hand.round_cards {
            for playing in rounds.iter() {
                let col = gtk::Box::new(gtk::Orientation::Vertical, 4);
                let pname = snapshot.players.as_ref()
                    .and_then(|p| p.iter().find(|pl| pl.id == playing.player_id))
                    .map(|pl| pl.name.as_str()).unwrap_or("?");
                
                let is_winner = winning_card_id.as_ref().map(|id| *id == format!("{}-{}-{}", playing.player_id, playing.card.rank, playing.card.suit)).unwrap_or(false);
                
                let plbl = gtk::Label::new(Some(pname));
                plbl.add_css_class("playing-nametag");
                if is_winner {
                    plbl.add_css_class("playing-nametag-winner");
                }
                
                col.append(&plbl);
                
                let c_widget = create_card_widget(Some(&playing.card));
                if is_winner {
                    c_widget.add_css_class("card-winner");
                }
                
                col.append(&c_widget);
                played_box.append(&col);
            }
        }
        center_box.append(&played_box);
        
        let manilha_stack = gtk::Box::new(gtk::Orientation::Vertical, 8);
        manilha_stack.append(&gtk::Label::new(Some("MANILHA")));
        if let Some(m) = &hand.manilha {
            let m_lbl = gtk::Label::new(None);
            m_lbl.set_size_request(86, 124);
            m_lbl.set_markup(&format!("<span size='34000'><b>{}</b></span>", m));
            m_lbl.add_css_class("manilha-area");
            m_lbl.add_css_class("manilha-glow");
            manilha_stack.append(&m_lbl);
        }
        center_box.append(&manilha_stack);
    }
    
    // My Hand and Controls
    let bottom_box = window.bottom_box();
    clear_box(&bottom_box);

    let actions = bundle.ui.as_ref().and_then(|u| u.actions.as_ref());
    let can_ask = actions.map(|a| a.can_ask_or_raise).unwrap_or(snapshot.can_ask_truco.unwrap_or(false));
    let must_respond = actions.map(|a| a.must_respond).unwrap_or(false);
    let can_play_card = actions.map(|a| a.can_play_card).unwrap_or(snapshot.turn_player == Some(local_idx));
    let is_my_turn = can_play_card || must_respond;
    let match_over = snapshot.match_finished.unwrap_or(false);
    
    if !match_over {
        let action_box = gtk::Box::new(gtk::Orientation::Horizontal, 16);
        action_box.set_halign(gtk::Align::Center);
        
        if must_respond {
            let btn_accept = gtk::Button::with_label("ACEITAR");
            btn_accept.add_css_class("btn-accept");
            let core_a = core.clone();
            btn_accept.connect_clicked(move |_| { dispatch_game_action(&core_a, "accept", None); });
            action_box.append(&btn_accept);
            
            let current_stake = snapshot.current_hand.as_ref().and_then(|h| h.stake).unwrap_or(1);
            if current_stake < 9 {
                let label = raise_label_for(next_stake(current_stake));
                let btn_raise = gtk::Button::with_label(&label);
                btn_raise.add_css_class("btn-truco");
                let core_r = core.clone();
                btn_raise.connect_clicked(move |_| { dispatch_game_action(&core_r, "truco", None); });
                action_box.append(&btn_raise);
            }
            
            let btn_refuse = gtk::Button::with_label("CORRER");
            btn_refuse.add_css_class("btn-refuse");
            let core_f = core.clone();
            btn_refuse.connect_clicked(move |_| { dispatch_game_action(&core_f, "refuse", None); });
            action_box.append(&btn_refuse);
        } else if can_ask {
            let current_stake = snapshot.current_hand.as_ref().and_then(|h| h.stake).unwrap_or(1);
            let label = raise_label_for(next_stake(current_stake));
            let btn_truco = gtk::Button::with_label(&label);
            btn_truco.add_css_class("btn-truco");
            let core_t = core.clone();
            btn_truco.connect_clicked(move |_| { dispatch_game_action(&core_t, "truco", None); });
            action_box.append(&btn_truco);
        }
        
        bottom_box.append(&action_box);
    }

    if is_my_turn && !match_over {
        let lbl = gtk::Label::new(Some("SUA VEZ"));
        lbl.add_css_class("turn-pill");
        bottom_box.append(&lbl);
    }

    if let Some(players) = &snapshot.players {
        if let Some(me) = players.iter().find(|p| p.id == local_idx) {
            // Name + Team badge
            let name_row = gtk::Box::new(gtk::Orientation::Horizontal, 8);
            name_row.set_halign(gtk::Align::Center);
            name_row.append(&gtk::Label::new(Some(&me.name.to_uppercase())));
            let team_lbl = gtk::Label::new(Some(&format!("Time {}", me.team + 1)));
            team_lbl.add_css_class(if me.team == 0 { "team-badge-us" } else { "team-badge-them" });
            name_row.append(&team_lbl);
            bottom_box.append(&name_row);
            
            let my_hand = gtk::Box::new(gtk::Orientation::Horizontal, 16);
            if let Some(cards) = &me.hand {
                for (idx, c) in cards.iter().enumerate() {
                    let card_widget = create_card_widget(Some(c));
                    card_widget.add_css_class("card-clickable");
                    if can_play_card {
                        let gesture = gtk::GestureClick::new();
                        let core_clone = core.clone();
                        gesture.connect_pressed(move |_, _, _, _| {
                            dispatch_game_action(&core_clone, "play", Some(idx));
                        });
                        card_widget.add_controller(gesture);
                    }
                    my_hand.append(&card_widget);
                }
            }
            bottom_box.append(&my_hand);
        }
    }
}

fn render_opponent(container: &gtk::Box, player: &models::Player, is_turn: bool, vertical_cards: bool, snap: &models::GameSnapshot) {
    // Role badge
    let badge = role_badge_for(player.id, snap);
    let mut name_text = String::new();
    if let Some(b) = &badge {
        name_text.push_str(b);
        name_text.push(' ');
    }
    name_text.push_str(&player.name.to_uppercase());
    if player.cpu == Some(true) {
        if is_turn {
            name_text.push_str(" ⟳"); // CPU spinner indicator
        } else {
            name_text.push_str(if player.provisional_cpu == Some(true) { " 🤖*" } else { " 🤖" });
        }
    }
    if is_turn && player.cpu != Some(true) {
        name_text.push_str(" ◀ VEZ");
    }
    let name_lbl = gtk::Label::new(Some(&name_text));
    name_lbl.add_css_class(if is_turn { "opponent-pill-active" } else { "opponent-pill" });
    container.append(&name_lbl);
    
    // Team badge
    let team_lbl = gtk::Label::new(Some(&format!("Time {}", player.team + 1)));
    team_lbl.add_css_class(if player.team == 0 { "team-badge-us" } else { "team-badge-them" });
    container.append(&team_lbl);
    
    let count = player.hand.as_ref().map(|h| h.len()).unwrap_or(3);
    let hand_box = gtk::Box::new(
        if vertical_cards { gtk::Orientation::Vertical } else { gtk::Orientation::Horizontal },
        if vertical_cards { -30 } else { -20 },
    );
    for _ in 0..count {
        hand_box.append(&create_card_widget(None));
    }
    container.append(&hand_box);
}

fn role_badge_for(player_id: i32, snap: &models::GameSnapshot) -> Option<String> {
    let hand = snap.current_hand.as_ref()?;
    let n = snap.num_players.unwrap_or(2);
    let dealer = hand.dealer?;
    if player_id == dealer { return Some("🃏".to_string()); }
    if player_id == (dealer + 1) % n { return Some("✋".to_string()); }
    if n == 4 && player_id == (dealer + n - 1) % n { return Some("🦶".to_string()); }
    None
}

fn next_stake(s: i32) -> i32 {
    match s { 1 => 3, 3 => 6, 6 => 9, 9 => 12, _ => s }
}

fn raise_label_for(stake: i32) -> String {
    match stake { 3 => "TRUCO!", 6 => "SEIS!", 9 => "NOVE!", 12 => "DOZE!", _ => "TRUCO!" }.to_string()
}

fn create_card_widget(card_opt: Option<&models::Card>) -> gtk::Box {
    let container = gtk::Box::new(gtk::Orientation::Vertical, 0);
    container.set_size_request(86, 124);
    
    if let Some(card) = card_opt {
        container.add_css_class("card");
        if card.is_red() { container.add_css_class("card-red"); }
        
        let top_str = format!("<span size='large'><b>{}</b></span>\n{}", card.rank, card.suit_symbol());
        let top_lbl = gtk::Label::new(None);
        top_lbl.set_markup(&top_str);
        top_lbl.set_halign(gtk::Align::Start);
        top_lbl.set_valign(gtk::Align::Start);
        top_lbl.set_margin_top(8);
        top_lbl.set_margin_start(8);
        
        let center_lbl = gtk::Label::new(None);
        center_lbl.set_markup(&format!("<span size='32000'><b>{}</b></span>", card.suit_symbol()));
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
